from src.core.config import get_settings

s = get_settings()

def main():
    print('OK: ', s.APP_NAME, s.async_database_url)


if __name__ == "__main__":
    main()
