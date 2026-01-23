#!/usr/bin/env python3
"""
Consumer-2: A CLI tool that uses the Calculator API for add, multiply, and divide operations.

Usage:
    python main.py add <x> <y> [--validate]
    python main.py multiply <x> <y> [--validate]
    python main.py divide <x> <y> [--validate]

Options:
    --validate  Enable CVT contract validation (default: off)
"""

import argparse
import os
import sys

import requests

# Configuration
PRODUCER_URL = os.getenv("PRODUCER_URL", "http://localhost:10001")
CVT_SERVER_ADDR = os.getenv("CVT_SERVER_ADDR", "localhost:9550")
SCHEMA_PATH = os.getenv("SCHEMA_PATH", "./calculator-api.yaml")


def create_session(validate: bool) -> requests.Session:
    """Creates a requests session, optionally wrapped with CVT validation."""
    if validate:
        try:
            from cvt_sdk import ContractValidator
            from cvt_sdk.adapters import ContractValidatingSession

            validator = ContractValidator(CVT_SERVER_ADDR)
            validator.register_schema("calculator-api", SCHEMA_PATH)

            session = ContractValidatingSession(
                validator,
                auto_validate=True,
                on_validation_failure=lambda result,
                req,
                resp: validation_failure_handler(result),
            )
            print("CVT validation enabled", file=sys.stderr)
            return session
        except ImportError:
            print(
                "Warning: CVT SDK not installed. Run with --validate requires cvt-sdk.",
                file=sys.stderr,
            )
            print("Continuing without validation...", file=sys.stderr)
        except Exception as e:
            print(f"Warning: Failed to enable CVT validation: {e}", file=sys.stderr)
            print("Continuing without validation...", file=sys.stderr)

    return requests.Session()


def validation_failure_handler(result: dict) -> None:
    """Handles CVT validation failures."""
    errors = ", ".join(result.get("errors", []))
    print(f"CVT Validation failed: {errors}", file=sys.stderr)
    sys.exit(1)


def handle_error(error: Exception) -> None:
    """Handles errors from API calls."""
    if isinstance(error, requests.exceptions.HTTPError):
        response = error.response
        if response is not None:
            try:
                data = response.json()
                if "error" in data:
                    print(f"Error: {data['error']}", file=sys.stderr)
                    sys.exit(1)
            except ValueError:
                pass
            print(f"Error: {response.status_code} {response.reason}", file=sys.stderr)
        else:
            print(f"Error: {error}", file=sys.stderr)
    elif isinstance(error, requests.exceptions.ConnectionError):
        print(
            "Error: No response from server. Is the producer running?", file=sys.stderr
        )
    else:
        print(f"Error: {error}", file=sys.stderr)
    sys.exit(1)


def add(x: float, y: float, validate: bool) -> None:
    """Performs an add operation."""
    try:
        session = create_session(validate)
        response = session.get(f"{PRODUCER_URL}/add", params={"x": x, "y": y})
        response.raise_for_status()
        result = response.json()["result"]
        print(f"{x} + {y} = {result}")
    except Exception as e:
        handle_error(e)


def multiply(x: float, y: float, validate: bool) -> None:
    """Performs a multiply operation."""
    try:
        session = create_session(validate)
        response = session.get(f"{PRODUCER_URL}/multiply", params={"x": x, "y": y})
        response.raise_for_status()
        result = response.json()["result"]
        print(f"{x} * {y} = {result}")
    except Exception as e:
        handle_error(e)


def divide(x: float, y: float, validate: bool) -> None:
    """Performs a divide operation."""
    try:
        session = create_session(validate)
        response = session.get(f"{PRODUCER_URL}/divide", params={"x": x, "y": y})
        response.raise_for_status()
        result = response.json()["result"]
        print(f"{x} / {y} = {result}")
    except Exception as e:
        handle_error(e)


def main() -> None:
    """Main entry point for the CLI."""
    parser = argparse.ArgumentParser(
        description="CLI tool for add, multiply, and divide operations using the Calculator API"
    )
    parser.add_argument("--version", action="version", version="consumer-2 1.0.0")

    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Add command
    add_parser = subparsers.add_parser("add", help="Add two numbers")
    add_parser.add_argument("x", type=float, help="First number")
    add_parser.add_argument("y", type=float, help="Second number")
    add_parser.add_argument(
        "--validate", action="store_true", help="Enable CVT contract validation"
    )

    # Multiply command
    multiply_parser = subparsers.add_parser("multiply", help="Multiply two numbers")
    multiply_parser.add_argument("x", type=float, help="First number")
    multiply_parser.add_argument("y", type=float, help="Second number")
    multiply_parser.add_argument(
        "--validate", action="store_true", help="Enable CVT contract validation"
    )

    # Divide command
    divide_parser = subparsers.add_parser("divide", help="Divide two numbers")
    divide_parser.add_argument("x", type=float, help="First number (dividend)")
    divide_parser.add_argument("y", type=float, help="Second number (divisor)")
    divide_parser.add_argument(
        "--validate", action="store_true", help="Enable CVT contract validation"
    )

    args = parser.parse_args()

    if args.command is None:
        parser.print_help()
        sys.exit(1)

    if args.command == "add":
        add(args.x, args.y, args.validate)
    elif args.command == "multiply":
        multiply(args.x, args.y, args.validate)
    elif args.command == "divide":
        divide(args.x, args.y, args.validate)


if __name__ == "__main__":
    main()
