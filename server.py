'''
A mockup server in Python to support light traffic, for quich verification of features, and API.
'''

import oauth2
from functools import wraps
from flask import Flask, request, abort
import hmac
import hashlib
# import time
from werkzeug.exceptions import Unauthorized

app = Flask(__name__)

oauth_server = oauth2.Server(signature_methods={
            # Supported signature methods
            'HMAC-SHA1': oauth2.SignatureMethod_HMAC_SHA1()
        })

C2S_SECRET_KEY = "XXXXXXXX",
S2S_SECRET_KEY = "XXXXXXXX"


def validate_two_leg_oauth():
    """
    Verify 2-legged oauth request. Parameters accepted as
    values in "Authorization" header, or as a GET request
    or in a POST body.
    """
    auth_header = {}
    if 'Authorization' in request.headers:
        auth_header = {'Authorization': request.headers['Authorization']}

    req = oauth2.Request.from_request(
        request.method,
        request.url,
        headers=auth_header,
        # the immutable type of "request.values" prevents us from sending
        # that directly, so instead we have to turn it into a python
        # dict
        parameters=dict([(k, v) for k, v in request.values.iteritems()]))

    print request.headers["Authorization"]

    try:
        oauth_server.verify_request(req,
            _get_consumer(request.values.get('oauth_consumer_key')),
            None)
        return True

    except oauth2.Error, e:
        raise Unauthorized(e)
    except KeyError, e:
        raise Unauthorized("You failed to supply the "\
                           "necessary parameters (%s) to "\
                           "properly authenticate " % e)


class MockConsumer(object):
    key = '-------'
    secret = '--------------'


def _get_consumer(key):
    """
    in real life we'd fetch a consumer object,
    using the provided key, that
    has at the bare minimum the attributes
    key and secret.
    """
    return MockConsumer()


def oauth_protect(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        validate_two_leg_oauth()
        return f(*args, **kwargs)
    return decorated_function


@app.route("/")
def hello():
    return "I'm exposed for everyone to see"


@app.route('/oauth', methods=['GET', 'POST'])
@oauth_protect
def private():
    return "I'm protected behind two-legged oauth"


# special simplified version of auth, much easier for internal service
@app.route('/simpleauth', methods=['GET', 'POST'])
def validate_simpleauth():
    authorization_type, raw_params = request.headers['authorization'].split(' ', 1)
    print authorization_type

    params = {}

    for param in raw_params.split(','):
        key, value = param.strip().split('=', 1)  # trim whitespace after comma
        params[key] = value[1:-1]  # remove quotes

    mac = hmac.new(S2S_SECRET_KEY if authorization_type == 'S2S' else C2S_SECRET_KEY, digestmod=hashlib.sha1)
    mac.update(request.method)
    mac.update(request.url)
    mac.update(request.data)  # body
    mac.update(params['timestamp'])

    expected_signature = mac.hexdigest()

    print "received ==> ", params['signature']
    print "expect ==> ", expected_signature

    if expected_signature != params['signature']:
        abort(401)  # unauthorized
        # return "wrong"
    else:
        return 'Hello, %s!' % params['realm']


if __name__ == '__main__':
    app.run(debug=True)
