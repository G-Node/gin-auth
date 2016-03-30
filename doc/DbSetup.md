GIN-Auth database setup
=======================

Create a database
-----------------

If necessary create a role and a database. Choose login name, password and database name as it fits
your needs.

```
sudo -u postgres psql -c "CREATE ROLE test WITH LOGIN PASSWORD 'test';"
sudo -u postgres psql -c "CREATE DATABASE gin_auth OWNER test;"
```

To connect to your database use the following command:

```
psql -W -U test -h localhost gin_auth
```

Apply database migrations
-------------------------

For database migrations GIN-Auth uses [goose](https://github.com/CloudCom/goose).

To apply/unapply all available migrations use the following commands:

```
goose -path ./conf up
goose -path ./conf down
```

The migration tool *goose* uses the configuration
file `conf/dbconf.yml`.
It might therefore be necessary to adapt the file to your environment before using the tool.
To learn more about *goose*, please read the [goose documentation](https://github.com/CloudCom/goose/blob/master/README.md).
