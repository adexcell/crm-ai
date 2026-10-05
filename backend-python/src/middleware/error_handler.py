import logging

from fastapi import FastAPI, Request, status
from fastapi.exceptions import HTTPException, RequestValidationError
from fastapi.responses import JSONResponse

from src.common.envelope import ApiResponse
from src.common.exceptions import AppException

logger = logging.getLogger("crm.api")

def setup_exception_handlers(app: FastAPI) -> None:
    @app.exception_handler(AppException)
    async def app_exception_handler(_: Request, exc: AppException) -> JSONResponse:
        envelope = ApiResponse.failure(
            code=exc.code,
            message=exc.message,
            details=exc.details,
        )
        return JSONResponse(
            status_code=exc.status_code, 
            content=envelope.model_dump(mode="json")
        )

    @app.exception_handler(RequestValidationError)
    async def validation_exception_handler(_: Request, exc: RequestValidationError) -> JSONResponse:
        errors = [
            {
                "field": ".".join(str(loc) for loc in err["loc"]),
                "message": err["msg"],
                "type": err["type"],
            }
            for err in exc.errors()
        ]
        envelope = ApiResponse.failure(
            code="VALIDATION_ERROR",
            message="Input validation failed",
            details={"errors": errors},
        )
        return JSONResponse(
            status_code=status.HTTP_422_UNPROCESSABLE_CONTENT,
            content=envelope.model_dump(mode="json"),
        )

    @app.exception_handler(Exception)
    async def unhandled_exception_handler(_: Request, exc: Exception) -> JSONResponse:
        logger.exception("Unhandled server exception: %s", exc)
        envelope = ApiResponse.failure(
            code="INTERNAL_SERVER_ERROR",
            message="An unexpected internal server error occurred",
        )
        return JSONResponse(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            content=envelope.model_dump(mode="json"),
        )

    @app.exception_handler(HTTPException)
    async def http_exception_handler(_: Request, exc: HTTPException) -> JSONResponse:
        envelope = ApiResponse.failure(
            code="HTTP_ERROR",
            message=exc.detail if isinstance(exc.detail, str) else "HTTP error",
            details={"status_code": exc.status_code},
        )
        return JSONResponse(
            status_code=exc.status_code, 
            content=envelope.model_dump(mode="json")
        )

