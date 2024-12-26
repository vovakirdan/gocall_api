# Secret key
You can also generate secret key manually (via python):
```python
import secrets
with open(".env", "a") as f:
    f.write(f"SECRET_KEY={secrets.token_hex(32)}")
```
Or just run the [script](generate_secret_key.py)
```bash
python3 generate_secret_key.py
```

# Run test api
> Don't forget to make it executable
```bash
chmod +x test_api.sh
```
Make sure you have installed [jq](https://stedolan.github.io/jq/)
```bash
sudo apt install jq
```
Run tests
```bash
./test_api.sh
```