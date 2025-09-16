import argparse
import json
import os
import logging
from pathlib import Path

from openeo.local import LocalConnection

logging.basicConfig(level=logging.INFO,
                    format='%(asctime)s - %(levelname)s - %(message)s',
                    handlers=[logging.StreamHandler()])

logger = logging.getLogger(__name__)

def execute_process_graph(pg_path: Path, output_path: Path):
    """Load a process graph JSON and execute it locally."""
    with open(pg_path) as f:
        process_graph = json.load(f)

    logger.info("Output will be saved to: %s", output_path)

    local_conn = LocalConnection("./")

    cube = local_conn.datacube_from_json(process_graph)

    result = local_conn.execute(cube)

    logger.info("Execution complete.")


def main():
    parser = argparse.ArgumentParser(description="Execute openEO process graph locally.")
    parser.add_argument("--pg", type=Path, required=True, help="Path to process graph JSON")
    parser.add_argument("--out", type=Path, required=True, help="Path to save results")

    args = parser.parse_args()

    if not args.pg.exists():
        raise FileNotFoundError(f"Process graph not found: {args.pg}")

    execute_process_graph(args.pg, args.out)


if __name__ == "__main__":
    main()
