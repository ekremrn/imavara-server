"""Load the project-owned JPEG fixtures through the ordinary contribution API."""
import argparse
import http.cookiejar
import json
import pathlib
import urllib.error
import urllib.request
import uuid

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("images", nargs="+", type=pathlib.Path)
    parser.add_argument("--base", default="http://127.0.0.1:18080/v1")
    args = parser.parse_args()
    session = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
    session.open(args.base + "/me").close()
    for image in args.images:
        boundary = uuid.uuid4().hex
        parts = [
            f'--{boundary}\r\nContent-Disposition: form-data; name="license_scope"\r\n\r\nshared\r\n'.encode(),
            f'--{boundary}\r\nContent-Disposition: form-data; name="image"; filename="fixture.jpg"\r\nContent-Type: image/jpeg\r\n\r\n'.encode(),
            image.read_bytes(),
            f'\r\n--{boundary}--\r\n'.encode(),
        ]
        request = urllib.request.Request(args.base + "/contributions", data=b"".join(parts),
            headers={"Content-Type": "multipart/form-data; boundary=" + boundary})
        try:
            with session.open(request, timeout=180) as response:
                result = json.load(response)
        except urllib.error.HTTPError as error:
            result = json.load(error)
            if error.code != 409 or result.get("status") != "duplicate":
                raise RuntimeError(f"{image.name}: {error.code} {result}") from error
        print(json.dumps({"filename": image.name, **result}, ensure_ascii=False), flush=True)

if __name__ == "__main__":
    main()
