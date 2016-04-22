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

##### Success

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

##### Success

If successful the response body contains the parameters `access_token`, `refresh_token` and `token_type` as JSON.

TODO should we also support other encodings (application/x-www-form-urlencoded) depending on the Accept header
of the request?

###Authenticate: grant type implicit

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

##### Success

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

###### Error

TODO

###### Success

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

##### Success

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

##### Success

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

##### Success

Submit configured scopes to [approve](#approve-scopes)

### Approve scopes

```
POST https://<host>/oauth/approve
```

##### Parameters (application/x-www-form-urlencoded)

| Name           | Type    | Description |
| -------------- | ------- | ---- |
| request_id     | string  | An id associated with a grant request (type code or implicit) |
| \<scope one\>  | bool    | True if scope one was granted, false otherwise |
| \<scope two\>  | bool    | True if scope two was granted, false otherwise |
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

##### Success

In case of success the browser is redirected to the `redirect_uri` associated with the respective request.
If the grant request type was code the redirect URL contains the parameters `code`, `scope` and `state`.
In case of an implicit grant request the redirect uri contains the parameters `access_token` and `token_type`.

### Validate tokens

TODO

### Account API

TODO

### SSH-key API

TODO
