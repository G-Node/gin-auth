GIN-Auth API
============

### Authenticate: grant type code


#### 1. Request access

The client should redirect the browser (302 moved temporarily) to the following URL
with the parameters listed below.

##### URL

```
GET https://<host>/oauth/authorize
```

##### Parameters

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| response_type | string  | Must be set to `code` |
| client_id     | string  | The ID of a registered client |
| redirect_uri  | string  | URL to redirect to after authorization |
| scope         | string  | Comma separated list of scopes |
| state         | string  | Random string to protect against CSRF |

##### Errors (not redirected)

Show an error page if:

* The client ID is unknown
* The redirect URL does not match exactly one registered URL for the client
* The redirect URL does not use https
* One of the given scopes is not registered

##### Response

Redirect the browser (302 moved temporarily) to the [login page](#login-page) using a newly generated request id.

#### 2. Exchange token

If the authentication and approval was successful the response is a redirect (302) to the requested `redirect_uri`
containing the parameters `code`, `scope` and `state`.
In the next step the access code can be used to obtain an access token.

##### URL

```
POST https://<host>/oauth/token
```

##### Basic authorization header

Send `client_id` and `client_secret` as HTTP basic authorization header.

##### Parameters (application/x-www-form-urlencoded)

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| code          | string  | The code obtained in step 1 |
| redirect_uri  | string  | URL to redirect to after authorization |
| grant_type    | string  | Must be 'authorization_code' |

##### Errors

Errors are returned encoded as json with the following format:

```javascript
{
  "code": 400,
  "error": "Bad Request",
  "message": "Unable to set one or more fields",
  "reasons": { // reasons may be null
    "grant_type": "Field grant type was missing"
  }
}
```

##### Response

If successful the response body contains the parameters `access_token`, `refresh_token` and `token_type` as JSON.

TODO should we also support other encodings (application/x-www-form-urlencoded) depending on the Accept header
of the request?

### Authenticate: grant type implicit

#### Request implicit access token

##### URL

```
GET https://<host>/oauth/authorize
```

##### Parameters

| Name          | Type    | Description |
| ------------- | ------- | ---- |
| response_type | string  | Must be set to `token` |
| client_id     | string  | The ID of a registered client |
| redirect_uri  | string  | URL to redirect to after authorization |
| scope         | string  | Comma separated list of scopes |
| state         | string  | Random string to protect against CSRF |

##### Errors (not redirected)

Show an error page if:

* The client ID is unknown
* The redirect URL does not match exactly one registered URL for the client
* The redirect URL does not use https
* One of the given scopes is not registered

##### Response

Redirect the browser (302 moved temporarily) to the [login page](#login-page) using a newly generated request id.

If the authentication and approval was successful the response is a redirect (302) to the requested `redirect_uri`
containing the parameters `access_token`, `token_type`, `scope` and `state`.

### Authenticate: grant type owner credentials

Non web applications can obtain a token by just using http basic auth.

#### Request access token directly

##### URL

```
POST https://<host>/oauth/token
```

##### Headers

| Name           | Value |
| -------------- | --- |
| Authorization  | Basic <username:password> |
| X-OAuth-Scopes | <list-of-scopes> |

##### Error

TODO

##### Response

If successful the response body contains the parameters `access_token` and `token_type` as application/x-www-form-urlencoded.

TODO should we also support other encodings (application/json) depending on the Accept header of the request?

### Login page

A login page shown during a `code` or `implicit` authorization request.

##### URL

```
GET https://<host>/oauth/login_page
```

##### Parameters

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

### Login

Checks the login credentials and the `request_id`

##### URL

```
POST https://<host>/oauth/login
```

##### Parameters (application/x-www-form-urlencoded)

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

### Approve page

Shows the resource owner a page that allows him to approve certain scopes for a specific client.

##### URL

```
GET https://<host>/oauth/approve_page
```

##### Parameters

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

### Approve scopes

```
POST https://<host>/oauth/approve
```

##### Parameters (application/x-www-form-urlencoded)

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

### Validate tokens

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
  "jti": "<token>",     // token identifier
  "exp": 1300819380,    // expiration time
  "iss": "gin-auth",
  "login": "...",       // login of the account
  "scope": ["s1", "s2"] // the scope which may be accessed with the token
}
```

### Account API

#### Error handling

In case of errors calls to the account API result in a response with the respective HTTP status code.
The body contains further information formatted as JSON:

```javascript
{
  "code": 400,
  "error": "Bad Request",
  "message": "...",
  "reasons": { // reasons may be null
    "foo": "Field foo was missing"
  }
}
```

#### List all accounts

##### URL

```
GET https://<host>/api/accounts
```

##### Authorization

A bearer token sent with the authorization header is required.
The token scope must contain 'account-admin'.

##### Response

Returns a list of all accounts as JSON:

```json
[
   {
       "url":  "https://<host>/api/accounts/<login>",
       "uuid": "...",
       "login": "<login>",
       "tile": "...",
       "first_name": "...",
       "middle_name": "...",
       "last_name": "...",
       "created_at": "YYYY-MM-DDThh:mm:ss",
       "updated_at": "YYYY-MM-DDThh:mm:ss"
   }
]
```

#### Get an account

##### URL

```
GET https://<host>/api/accounts/<login>
```

##### Authorization

A bearer token sent with the authorization header is required.
The token scope must contain 'account-read' to access own accounts or 'account-admin'.

##### Response

Returns an account object as JSON:

```json
{
   "url":  "https://<host>/api/accounts/<login>",
   "uuid": "...",
   "login": "<login>",
   "tile": "...",
   "first_name": "...",
   "middle_name": "...",
   "last_name": "...",
   "created_at": "YYYY-MM-DDThh:mm:ss",
   "updated_at": "YYYY-MM-DDThh:mm:ss"
}
```

#### Update an account

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
   "tile": "...",
   "first_name": "...",
   "middle_name": "...",
   "last_name": "..."
}
```

##### Response

The changed account object as JSON (see above).

#### Update account password

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

### SSH-key API

#### List keys per user

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

#### Get ssh key

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

#### Remove ssh key

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
