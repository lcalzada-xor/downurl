#!/usr/bin/env python3
import os
import sys
import requests
import hashlib
import tarfile
from urllib.parse import urlparse
from concurrent.futures import ThreadPoolExecutor, as_completed

WORKERS = 10

def filename_from_url(url):
    """Genera un nombre de archivo seguro desde un URL"""
    p = urlparse(url)
    name = p.path.rstrip("/").split("/")[-1]
    if not name or "." not in name:
        h = hashlib.sha1(url.encode()).hexdigest()[:10]
        ext = ".js" if url.endswith(".js") or url.endswith(".mjs") else ""
        return f"{h}{ext}"
    return "".join(c if c.isalnum() or c in "-_." else "_" for c in name)

def download_file(session, url, dest):
    """Descarga un archivo desde url a dest"""
    try:
        r = session.get(url, timeout=15)
        r.raise_for_status()
        os.makedirs(os.path.dirname(dest), exist_ok=True)
        with open(dest, "wb") as f:
            f.write(r.content)
        return True, dest
    except Exception as e:
        return False, str(e)

def main(url_file_or_dir):
    output_dir = "output"
    os.makedirs(output_dir, exist_ok=True)
    urls = []

    # Leer URLs desde archivo
    if os.path.isfile(url_file_or_dir):
        with open(url_file_or_dir, "r") as f:
            urls = [line.strip() for line in f if line.strip()]
    else:
        print(f"El argumento debe ser un archivo de URLs: {url_file_or_dir}")
        sys.exit(1)

    session = requests.Session()
    results = []

    for url in urls:
        result = {"url": url, "host": None, "downloaded": [], "errors": []}
        parsed = urlparse(url)
        host = parsed.netloc
        result["host"] = host
        js_dir = os.path.join(output_dir, host, "js")
        success, info = download_file(session, url, os.path.join(js_dir, filename_from_url(url)))
        if success:
            result["downloaded"].append(info)
        else:
            result["errors"].append(info)
        results.append(result)
        print(f"[*] Procesado {url} -> Descargados: {len(result['downloaded'])}, Errores: {len(result['errors'])}")

    # Guardar reporte
    report_path = os.path.join(output_dir, "report.txt")
    with open(report_path, "w") as f:
        for r in results:
            f.write(f"URL: {r['url']}\nHOST: {r['host']}\n")
            f.write(f"Downloaded: {len(r['downloaded'])}\n")
            for d in r['downloaded']:
                f.write(f"  - {d}\n")
            f.write(f"Errors: {len(r['errors'])}\n")
            for e in r['errors']:
                f.write(f"  - {e}\n")
            f.write("\n")

    # Empaquetar todo en tar.gz
    tar_path = os.path.join(output_dir, "output.tar.gz")
    with tarfile.open(tar_path, "w:gz") as tar:
        tar.add(output_dir, arcname=os.path.basename(output_dir))

    print(f"[+] Hecho. Salida en: {output_dir}/, informe: {report_path}, archivo: {tar_path}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(f"Uso: {sys.argv[0]} urls.txt")
        sys.exit(1)
    main(sys.argv[1])
