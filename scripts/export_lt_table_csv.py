#!/usr/bin/env python3
"""Export joined order and product data as a Lingtong table import CSV."""

from __future__ import annotations

import argparse
import csv
from datetime import datetime
from pathlib import Path
from typing import Any

import import_lt_table_xlsx as importer


def format_value(title: str, value: Any, *, product: bool = False) -> Any:
    if title in (importer.PRODUCT_DATE_HEADERS if product else importer.DATE_HEADERS):
        timestamp = importer.normalize_date(value)
        if timestamp is None:
            return ""
        return datetime.fromtimestamp(timestamp / 1000, importer.SHANGHAI).strftime(
            "%Y-%m-%d %H:%M:%S"
        )
    if title in (importer.PRODUCT_NUMBER_HEADERS if product else importer.NUMBER_HEADERS):
        value = importer.normalize_number(value)
        return "" if value is None else value
    return importer.normalize_text(value)


def export_csv(args: argparse.Namespace) -> int:
    source_dir = args.source_dir.expanduser().resolve()
    product_file = args.product_file.expanduser().resolve()
    output = args.output.expanduser().resolve()

    required_codes = importer.collect_order_product_codes(source_dir)
    product_lookup = importer.load_product_lookup(product_file, required_codes)
    product_headers = [
        title
        for title in importer.PRODUCT_HEADERS
        if title not in importer.PRODUCT_EXCLUDED_HEADERS
    ]
    headers = [
        importer.IMPORT_KEY_TITLE,
        *importer.SOURCE_HEADERS,
        *(f"{title}(档案)" for title in product_headers),
    ]

    output.parent.mkdir(parents=True, exist_ok=True)
    exported = 0
    last_sequence = args.start_after
    with output.open("w", encoding="utf-8", newline="") as stream:
        writer = csv.writer(stream)
        writer.writerow(headers)
        for row in importer.iter_source_rows(source_dir):
            sequence = importer.normalize_sequence(row[0])
            if sequence <= args.start_after:
                continue
            product = product_lookup[importer.normalize_text(row[5])]
            writer.writerow(
                [f"{args.import_batch}:{sequence}"]
                + [
                    format_value(title, row[index])
                    for index, title in enumerate(importer.SOURCE_HEADERS)
                ]
                + [
                    format_value(title, product.get(title), product=True)
                    for title in product_headers
                ]
            )
            exported += 1
            last_sequence = sequence
            if exported >= args.limit:
                break

    print(
        f"已导出 {exported:,} 条到 {output}，"
        f"序号范围 {args.start_after + 1:,}-{last_sequence:,}"
    )
    return 0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-dir", type=Path, required=True)
    parser.add_argument("--product-file", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--import-batch", required=True)
    parser.add_argument("--start-after", type=int, default=0)
    parser.add_argument("--limit", type=int, required=True)
    args = parser.parse_args()
    if args.start_after < 0:
        parser.error("--start-after 不能小于 0")
    if args.limit <= 0:
        parser.error("--limit 必须大于 0")
    return args


if __name__ == "__main__":
    raise SystemExit(export_csv(parse_args()))
