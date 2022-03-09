import sign384
import sys

api_secret = sys.argv[1]
request_body = sys.argv[2]

result = sign384.sign_api_request(api_secret, request_body)

print(result)
