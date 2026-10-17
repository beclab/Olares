"""Regression checks for Market listing preflight, without external network."""
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch, MagicMock

SOURCE = Path(__file__).parent / "olares-publish/scripts/check_listing.py"
spec = importlib.util.spec_from_file_location("check_listing", SOURCE)
listing = importlib.util.module_from_spec(spec)
spec.loader.exec_module(listing)


class ListingTests(unittest.TestCase):
    def test_collects_all_declared_images(self):
        data = {"metadata": {"icon": "https://example.test/icon"},
                "spec": {"featuredImage": "https://example.test/hero",
                         "promoteImage": ["https://example.test/one", "https://example.test/two"]},
                "entrances": [{"icon": "https://example.test/entrance"}],
                "containers": [{"image": "registry/container:tag"}]}
        self.assertEqual(len(list(listing.image_urls(data))), 5)

    def test_encoding_regressions_and_valid_locales(self):
        for text, expected in [("中文、日本語、français — für", 0),
                               ("功能�?", 1), ("cafÃ© â€™", 1)]:
            with self.subTest(text=text), tempfile.TemporaryDirectory() as temp:
                chart = Path(temp)
                (chart / "OlaresManifest.yaml").write_text("metadata:\n  title: " + text, encoding="utf-8")
                self.assertEqual(listing.check_chart(chart), expected)

    def test_invalid_utf8(self):
        with tempfile.TemporaryDirectory() as temp:
            chart = Path(temp)
            (chart / "OlaresManifest.yaml").write_bytes(b"metadata: \xff")
            self.assertEqual(listing.check_chart(chart), 1)

    def test_locales_deduplicate_urls_and_detect_escaped_corruption(self):
        with tempfile.TemporaryDirectory() as temp, patch.object(listing, "check_url") as check:
            chart = Path(temp)
            locale = chart / "i18n/zh-CN"
            locale.mkdir(parents=True)
            content = 'metadata:\n  icon: https://example.test/icon\n'
            (chart / "OlaresManifest.yaml").write_text(content, encoding="utf-8")
            (locale / "OlaresManifest.yaml").write_text(content + '  title: "\\uFFFD"\n', encoding="utf-8")
            self.assertEqual(listing.check_chart(chart), 1)
            check.assert_called_once()

    def test_get_image_response_and_failures(self):
        with patch.object(listing, "urlopen") as open_url:
            response = MagicMock()
            open_url.return_value.__enter__.return_value = response
            response.status = 200
            response.headers.get_content_type.return_value = "image/png"
            response.read.return_value = b"image bytes"
            listing.check_url("https://example.test/icon", 1)
            self.assertEqual(open_url.call_args[0][0].get_method(), "GET")
            response.headers.get_content_type.return_value = "text/html"
            with self.assertRaises(ValueError):
                listing.check_url("https://example.test/icon", 1)
            response.headers.get_content_type.return_value = "image/png"
            response.read.return_value = b""
            with self.assertRaises(ValueError):
                listing.check_url("https://example.test/icon", 1)
            open_url.side_effect = TimeoutError("timeout")
            with self.assertRaises(TimeoutError):
                listing.check_url("https://example.test/icon", 1)


if __name__ == "__main__":
    unittest.main()
