import logging

from rich.logging import RichHandler


def setup_logging(level: int = logging.INFO) -> None:
    logging.basicConfig(
        level=level,
        format="%(message)s",
        datefmt="[%X]",
        handlers=[
            RichHandler(
                rich_tracebacks=True,
                tracebacks_show_locals=False,
                tracebacks_suppress=[
                    "uvicorn",
                    "starlette",
                    "fastapi",
                    "pydantic",
                    "anyio",
                ],
                show_path=False,
                omit_repeated_times=False,
                markup=True,
            )
        ],
    )
