import os
from pathlib import Path
import subprocess
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("retry.sh")
NETWORK_ERROR = (
    'Get "https://proxy.golang.org/example.com/module/@v/v1.0.0.zip": '
    "stream error: stream ID 37; INTERNAL_ERROR; received from peer"
)


class RetryTest(unittest.TestCase):
    def run_command(self, body, **overrides):
        with tempfile.TemporaryDirectory() as directory:
            counter = Path(directory) / "attempts"
            env = dict(os.environ, RETRY_ATTEMPTS="3", RETRY_DELAY="0")
            env.update(overrides)
            env["COUNTER"] = str(counter)
            env["NETWORK_ERROR"] = NETWORK_ERROR
            env["RETRY_COMMAND"] = 'echo attempt >> "$COUNTER"\n' + body
            result = subprocess.run(
                ["bash", str(SCRIPT)], env=env, capture_output=True, text=True
            )
            attempts = len(counter.read_text().splitlines()) if counter.exists() else 0
            return result.returncode, attempts, result.stdout + result.stderr

    def test_success_runs_once(self):
        status, attempts, _ = self.run_command("true")
        self.assertEqual((status, attempts), (0, 1))

    def test_transient_failure_recovers(self):
        status, attempts, _ = self.run_command(
            'if [ "$(wc -l < "$COUNTER")" -lt 3 ]; then\n'
            '  echo "$NETWORK_ERROR" >&2\n'
            '  exit 7\n'
            'fi'
        )
        self.assertEqual((status, attempts), (0, 3))

    def test_persistent_failure_preserves_exit_status(self):
        status, attempts, _ = self.run_command('echo "$NETWORK_ERROR" >&2; exit 7')
        self.assertEqual((status, attempts), (7, 3))

    def test_non_transient_errors_are_not_retried(self):
        for error in (
            "undefined: missingSymbol",
            "checksum mismatch",
            'https://proxy.golang.org/example.com/@v/v1.zip: 404 Not Found',
            'https://example.com/file: INTERNAL_ERROR',
        ):
            with self.subTest(error=error):
                status, attempts, _ = self.run_command(
                    'printf "%s\\n" "$ERROR" >&2; exit 9', ERROR=error
                )
                self.assertEqual((status, attempts), (9, 1))

    def test_command_pipeline_failure_is_preserved(self):
        status, attempts, _ = self.run_command("false | cat")
        self.assertEqual((status, attempts), (1, 1))

    def test_cancellation_is_not_retried(self):
        for code in (130, 143):
            with self.subTest(code=code):
                status, attempts, _ = self.run_command(
                    f'echo "$NETWORK_ERROR" >&2; exit {code}'
                )
                self.assertEqual((status, attempts), (code, 1))

    def test_invalid_limits_rejected_before_execution(self):
        for overrides in (
            {"RETRY_ATTEMPTS": "0"},
            {"RETRY_ATTEMPTS": "abc"},
            {"RETRY_DELAY": "-1"},
            {"RETRY_DELAY": "08"},
        ):
            with self.subTest(overrides=overrides):
                status, attempts, _ = self.run_command("true", **overrides)
                self.assertEqual((status, attempts), (2, 0))


if __name__ == "__main__":
    unittest.main()
