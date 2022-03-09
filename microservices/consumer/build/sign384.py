import hashlib
import hmac

def sign_api_request(api_secret, requestbody):
    signature_hash = hmac.new(api_secret.encode(), requestbody.encode(),  hashlib.sha384).hexdigest()
    return signature_hash
