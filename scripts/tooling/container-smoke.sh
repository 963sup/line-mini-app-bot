#!/usr/bin/env bash
# Verify image contents without credentials, network or LINE publication.
set -euo pipefail
image="${1:?Usage: container-smoke.sh IMAGE}"
[[ $(docker image inspect --format '{{.Config.User}}' "$image") == '65532:65532' ]]
[[ $(docker image inspect --format '{{json .Config.Entrypoint}}' "$image") == '["/server"]' ]]
docker run --rm --network=none --read-only --cap-drop=ALL \
  --security-opt=no-new-privileges --entrypoint=/bin/sh "$image" -ceu '
  test -x /server
  test -r /app/frontend/miniapp.html || test -r /app/frontend/miniapp/index.html
  test -r /assets/line/richmenu/attendance-in.jpg
  test -r /assets/line/richmenu/attendance-out.jpg
  test -s /etc/ssl/certs/ca-certificates.crt
  pdftotext -v
  # An invalid argument must be rejected before any external side effect.
  set +e
  output=$(/server __container_smoke_invalid_mode__ 2>&1)
  code=$?
  set -e
  test "$code" -eq 2
  test "$output" = "unsupported application mode"
  set +e
  output=$(/server sync-line-navigation __unexpected_argument__ 2>&1)
  code=$?
  set -e
  test "$code" -eq 2
  test "$output" = "sync-line-navigation takes no arguments"
'
# Exercise PDF extraction through stdin/stdout with a generated one-page fixture.
python3 - "$image" <<'PYTHON'
import subprocess, sys
stream = b"BT /F1 12 Tf 20 100 Td (container-smoke) Tj ET"
objects = [
    b"<< /Type /Catalog /Pages 2 0 R >>",
    b"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
    b"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
    b"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
    b"<< /Length " + str(len(stream)).encode() + b" >>\nstream\n" + stream + b"\nendstream",
]
pdf = bytearray(b"%PDF-1.4\n")
offsets = [0]
for i, obj in enumerate(objects, 1):
    offsets.append(len(pdf))
    pdf.extend(f"{i} 0 obj\n".encode() + obj + b"\nendobj\n")
xref = len(pdf)
pdf.extend(f"xref\n0 {len(offsets)}\n0000000000 65535 f \n".encode())
for offset in offsets[1:]:
    pdf.extend(f"{offset:010d} 00000 n \n".encode())
pdf.extend(f"trailer\n<< /Size {len(offsets)} /Root 1 0 R >>\nstartxref\n{xref}\n%%EOF\n".encode())
result = subprocess.run([
    "docker", "run", "--rm", "-i", "--network=none", "--read-only",
    "--cap-drop=ALL", "--security-opt=no-new-privileges",
    "--entrypoint=pdftotext", sys.argv[1], "-layout", "-enc", "UTF-8", "-", "-",
], input=pdf, capture_output=True, check=True, timeout=30)
assert b"container-smoke" in result.stdout, "PDF text extraction failed"
PYTHON
echo 'PASS: non-root image, entrypoint, static assets, CA, Poppler and CLI dispatch'
