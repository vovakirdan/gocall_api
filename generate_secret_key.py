import secrets
with open(".env", "a") as f:
    f.write(f"SECRET_KEY={secrets.token_hex(32)}")