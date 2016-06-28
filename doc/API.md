GIN-Auth API
============



Authenticate: grant type code
-----------------------------

### 1. Request access

The client should redirect the browser (302) to the following URL with the parameters listed below encoded
in the query string.

##### URL

```
GET https://<host>/oauth/authorize
```

##### Query Parameters

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| response_type | string  | Must be set to `code` |
| client_id     | string  | The ID of a registered client |
| redirect_uri  | string  | URL to redirect to after authorization |
| scope         | string  | Space separated list of scopes |
| state         | string  | Random string to protect against CSRF |

##### Errors

Show an error page if:

* The client ID is unknown
* The redirect URL does not match exactly one registered URL for the client
* The redirect URL does not use https
* One of the given scopes is not registered or blacklisted

##### Response

Redirect the browser (302) to a page which performs an appropriate authentication and approval process.
If the authentication and approval was successful the response is a redirect (302) to the requested
`redirect_uri` containing the parameters `code`, `scope` and `state` as query parameters.

In the next step the `code` can be exchanged for an access and refresh token.

### 2. Exchange an access code for a token

To exchange a previously issued code for an access token.
The client must provide its `client_id` and `client_secret` either with the `Authorization` header or encoded in
the request body.

##### URL

```
POST https://<host>/oauth/token
```

##### Authorization Header

Send `client_id` and `client_secret` as HTTP basic authorization header (optional).

##### Request Body (application/x-www-form-urlencoded)

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| code          | string  | The code obtained in step 1 |
| grant_type    | string  | Must be 'authorization_code' |
| client_id     | string  | The client id (optional if the authorization header is present) |
| client_secret | string  | The client secret (optional if the authorization header is present) |

##### Errors

Return an error if:

* The client ID is unknown
* The client secret does not match
* The code is not valid

Errors are returned encoded as JSON using the following format:

```json
{
  "code": 400,
  "error": "Bad Request",
  "message": "Unable to set one or more fields",
  "reasons": {
    "grant_type": "Field grant type was missing"
  }
}
```

##### Response

If successful the response body contains the `scope`, `access_token`, `refresh_token` and `token_type`
as JSON.

```json
{
  "scope": "scope1 scope2",
  "access_token": "...",
  "refresh_token": "...",
  "token_type": "Bearer"
}
```

*TODO: should we also support other encodings (application/x-www-form-urlencoded) depending on the Accept header
of the request?*

### 3. Exchange a refresh token for an access token

To exchange a previously issued refresh token for an access token, the client must provide its `client_id` and `client_secret`
either with the `Authorization` header or encoded in the request body.

##### URL

```
POST https://<host>/oauth/token
```

##### Authorization Header

Send `client_id` and `client_secret` as HTTP basic authorization header (optional).

##### Request Body (application/x-www-form-urlencoded)

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| grant_type    | string  | Must be 'refresh_token' |
| refresh_token | string  | The refresh token |
| client_id     | string  | The client id (optional if the authorization header is present) |
| client_secret | string  | The client secret (optional if the authorization header is present) |

##### Errors

Return an error if:

* The client ID is unknown
* The client secret does not match
* The refresh token is not valid for the client

Errors are returned encoded as JSON in the [above shown format](#errors-1).

##### Response

If successful the response body contains the parameters `scope`, `access_token` and `token_type` as JSON.

```json
{
  "scope": "scope1 scope2",
  "access_token": "...",
  "refresh_token": "...",
  "token_type": "Bearer"
}
```



Authenticate: grant type implicit
---------------------------------

##### URL

```
GET https://<host>/oauth/authorize
```

##### Query Parameters

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| response_type | string  | Must be set to `token` |
| client_id     | string  | The ID of a registered client |
| redirect_uri  | string  | URL to redirect to after authorization |
| scope         | string  | Space separated list of scopes |
| state         | string  | Random string to protect against CSRF |

##### Errors (not redirected)

Show an error page if:

* The client ID is unknown
* The redirect URL does not match exactly one registered URL for the client
* The redirect URL does not use https
* One of the given scopes is not registered

##### Response

Redirect the browser (302) to a page which performs an appropriate authentication and approval process.

If the authentication and approval was successful the response is a redirect (302) to the requested `redirect_uri`
containing the parameters `access_token`, `token_type`, `scope` and `state`.



Authenticate: grant type owner credentials
------------------------------------------

Non web applications can obtain a token with the resource owners credentials.
To get an access token, the client must provide its `client_id` and `client_secret` either with the
`Authorization` header or encoded in the request body.

### Request access token directly

##### URL

```
POST https://<host>/oauth/token
```

##### Headers

Send `client_id` and `client_secret` as HTTP basic authorization header (optional).

##### Request Body (application/x-www-form-urlencoded)

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| scope         | string  | Space separated list of scopes |
| username      | string  | The resource owners login name |
| password      | string  | The resource owners password |
| grant_type    | string  | Must be 'password' |
| client_id     | string  | The client id (optional if the authorization header is present) |
| client_secret | string  | The client secret (optional if the authorization header is present) |

##### Errors

Return an error if:

* The client ID is unknown
* The client secret does not match
* The user credentials are not valid
* The requested scope is not whitelisted

Errors are returned encoded as JSON in the [above shown format](#errors-1).

##### Response

If successful the response body contains the parameters `scope`, `access_token` and `token_type` as JSON.

```json
{
  "scope": "scope1 scope2",
  "access_token": "...",
  "token_type": "Bearer"
}
```



Authenticate: grant type client credentials
-------------------------------------------

Clients can request an access token with limited privileges / scope directly with their client id and secret.
To get an access token, the client must provide its `client_id` and `client_secret` either with the
`Authorization` header or encoded in the request body.

### Request access token directly

##### URL

```
POST https://<host>/oauth/token
```

##### Headers

Send `client_id` and `client_secret` as HTTP basic authorization header (optional).

##### Request Body (application/x-www-form-urlencoded)

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| scope         | string  | Space separated list of scopes |
| grant_type    | string  | Must be 'client_credentials' |
| client_id     | string  | The client id (optional if the authorization header is present) |
| client_secret | string  | The client secret (optional if the authorization header is present) |

##### Errors

Return an error if:

* The client ID is unknown
* The client secret does not match
* The requested scope is not whitelisted

Errors are returned encoded as JSON in the [above shown format](#errors-1).

##### Response

If successful the response body contains the parameters `scope`, `access_token` and `token_type` as JSON.

```json
{
  "scope": "scope1 scope2",
  "access_token": "...",
  "token_type": "Bearer"
}
```



Login page
----------

A login page shown during a `code` or `implicit` authorization request.

##### URL

```
GET https://<host>/oauth/login_page
```

##### Query Parameters

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| request_id    | string  | An id associated with a grant request (type code or implicit) |

##### Cookies

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| session       | string  | A valid session cookie (optional) |

##### Errors

Show an error page if the the `request_id` does not belong to a registered request.

##### Response

If a valid `session` cookie is present the browser is redirected (302 moved temporarily) to the [approve page](#approve-page).
If the browser did not send a valid session the login form is shown and submitted to [login](#login)



Login
-----

Checks the login credentials and the `request_id`

##### URL

```
POST https://<host>/oauth/login
```

##### Request Body (application/x-www-form-urlencoded)

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| login         | string  | A unique user name |
| password      | string  | The users password |
| request_id    | string  | An id associated with a grant request (type code or implicit) |

##### Errors

Show an error page if the `request_id` does not match.

Show the form again if the credentials are not correct.

##### Response

If the parameters are accepted the response issues a session cookie called `session`.

If the user has already approved the client with all requested scopes the browser is redirected to the `redirect_uri`
associated with the respective request.
If the grant request type was code the redirect URL contains the parameters `code`, `scope` and `state`.
In case of an implicit grant request the redirect uri contains the parameters `access_token` and `token_type`.

If the user has not approved one of the requested scopes for this client before the browser is redirected to
the [approve page](#approve-page).



Approve page
------------

Shows the resource owner a page that allows him to approve certain scopes for a specific client.

##### URL

```
GET https://<host>/oauth/approve_page
```

##### Query Parameters

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| request_id    | string  | An id associated with a grant request (type code or implicit) |

##### Cookies

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| session       | string  | A valid session cookie |


##### Errors

Show an error page if the `request_id` is not valid.

Redirect the user to the login form if the `session` is not valid.

##### Response

Submit configured scopes to [approve](#approve-scopes)



Approve scopes
--------------

##### URL

```
POST https://<host>/oauth/approve
```

##### Request Body (application/x-www-form-urlencoded)

| Name           | Type    | Description |
| -------------- | ------- | ---- |
| request_id     | string  | An id associated with a grant request (type code or implicit) |
| scope          | string  | The first scope |
| scope          | string  | The second scope |
| ...            | ...     | ... |

##### Cookies

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| session       | string  | A valid session cookie |

##### Errors

Show an error page if:

* The `request_id` is not valid
* One of the scopes is not known

Redirect the user to the login form if the `session` is not valid.

##### Response

In case of success the browser is redirected to the `redirect_uri` associated with the respective request.
If the grant request type was code the redirect URL contains the parameters `code`, `scope` and `state`.
In case of an implicit grant request the redirect uri contains the parameters `access_token` and `token_type`.



Validate tokens
---------------

Validates an access token and provides information about the token such as expiration date and scope.

##### URL

```
GET https://<host>/oauth/validate/<token>
```

##### Errors

Return a json error (404 / Not Found) if the token does not exist or was expired.

##### Response

Returns information about the token encoded as JSON.

```javascript
{
  "url": "https://<host>/oauth/validate/<token>",
  "jti": "<token>",        // token identifier
  "exp": 1300819380,       // expiration time
  "iss": "gin-auth",
  "login": "...",          // login of the account (null if not not accociated with an account)
  "account_url": "...",    // url to the the account (null if not not accociated with an account)
  "scope": "scope1 scope2" // space separated list of scopes
}
```



Account API
-----------

### Error handling

In case of errors calls to the account API result in a response with the respective HTTP status code.
The body contains further information formatted as JSON:

```json
{
  "code": 400,
  "error": "Bad Request",
  "message": "...",
  "reasons": {
    "foo": "Field foo was missing"
  }
}
```

### Get an account

##### URL

```
GET https://<host>/api/accounts/<login>
```

##### Authorization

No authorization header required. However, to access non public `email` or `affiliation` information a
bearer token must sent with the authorization header.
The token scope must contain 'account-read' to access the own account or 'account-admin'.

##### Response

Returns a list of all accounts as JSON (depending on access restrictions `email` and/or `affiliation` may be null):

```json
{
   "url":  "https://<host>/api/accounts/<login>",
   "uuid": "...",
   "login": "<login>",
   "title": "...",
   "first_name": "...",
   "middle_name": "...",
   "last_name": "...",
   "email": {
       "email": "...",
       "is_public": true
   },
   "affiliation": {
       "institute": "...",
       "department": "...",
       "city": "...",
       "country": "...",
       "is_public": true
   },
   "created_at": "YYYY-MM-DDThh:mm:ss",
   "updated_at": "YYYY-MM-DDThh:mm:ss"
}
```

### List all accounts

##### URL

```
GET https://<host>/api/accounts
```

##### Query Parameters

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| q             | string  | A search string (optional) |

##### Authorization

No authorization header required. However, to access non public `email` or `affiliation` information a
bearer token must sent with the authorization header.

##### Response

Returns a list of all accounts as JSON in the above described format.

### Update an account

##### URL

```
PUT https://<host>/api/accounts/<login>
```

##### Authorization

A bearer token sent with the authorization header is required.
The token scope must contain 'account-write' to update own accounts or 'account-admin'.

##### Body

The request body must contain an account object with the following attributes.
Additional attributes may be present, but will be ignored.

```json
{
   "title": "...",
   "first_name": "...",
   "middle_name": "...",
   "last_name": "...",
   "email": {
      "email": "...",
      "is_public": true
  },
  "affiliation": {
      "institute": "...",
      "department": "...",
      "city": "...",
      "country": "...",
      "is_public": true
  }
}
```

##### Response

The changed account object as JSON (see above).

### Update account password

##### URL

```
PUT https://<host>/api/accounts/<login>/password
```

##### Authorization

A bearer token sent with the authorization header is required.
The token scope must contain 'account-write' to change the own password.

##### Body

```json
{
    "password_old": "...",
    "password_new": "...",
    "password_new_repeat": "..."
}
```

##### Response

If the password was successfully changed the status code is 200 and the response body is empty.



SSH-key API
-----------

### List keys per user

##### URL

```
GET https://<host>/api/accounts/<login>/keys
```

##### Authorization

A bearer token sent with the authorization header is required.
The token scope must contain 'account-read' to access own keys or 'account-admin'.

##### Response

Returns a list of ssh key objects as JSON:

```json
[
    {
        "url": "https://<host>/api/keys/<fingerprint>",
        "fingerprint": "<fingerprint>",
        "key": "...",
        "description": "...",
        "login": "<login>",
        "account_url": "https://<host>/api/accounts/<login>",
        "created_at": "YYYY-MM-DDThh:mm:ss",
        "updated_at": "YYYY-MM-DDThh:mm:ss"
    }
]
```

### Get ssh key

##### URL

```
GET https://<host>/api/keys/<fingerprint>
```

##### Authorization

A bearer token sent with the authorization header is required.
The token scope must contain 'account-read' to access own keys or 'account-admin'.

##### Response

Returns an ssh key object as JSON:

```json
{
    "url": "https://<host>/api/keys/<fingerprint>",
    "fingerprint": "<fingerprint>",
    "key": "...",
    "description": "...",
    "login": "<login>",
    "account_url": "https://<host>/api/accounts/<login>",
    "created_at": "YYYY-MM-DDThh:mm:ss",
    "updated_at": "YYYY-MM-DDThh:mm:ss"
}
```

### Remove ssh key

##### URL

```
DELETE https://<host>/api/keys/<fingerprint>
```

##### Authorization

A bearer token sent with the authorization header is required.
The token scope must contain 'account-write' to delete own keys.

##### Response

Returns the deleted ssh key object as JSON:

```json
{
    "url": "https://<host>/api/keys/<fingerprint>",
    "fingerprint": "<fingerprint>",
    "key": "...",
    "description": "...",
    "login": "<login>",
    "account_url": "https://<host>/api/accounts/<login>",
    "created_at": "YYYY-MM-DDThh:mm:ss",
    "updated_at": "YYYY-MM-DDThh:mm:ss"
}
```
