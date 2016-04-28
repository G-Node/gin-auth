-- Copyright (c) 2016, German Neuroinformatics Node (G-Node),
--                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
-- All rights reserved.
--
-- Redistribution and use in source and binary forms, with or without
-- modification, are permitted under the terms of the BSD License. See
-- LICENSE file in the root of the Project.


-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE Accounts (
  uuid              VARCHAR(36) PRIMARY KEY ,
  login             VARCHAR(512) NOT NULL UNIQUE ,
  email             VARCHAR(512) NOT NULL UNIQUE ,
  pwHash            VARCHAR(512) NOT NULL ,
  title             VARCHAR(512),
  firstName         VARCHAR(512) NOT NULL,
  middleName        VARCHAR(512),
  lastName          VARCHAR(512) NOT NULL,
  activationCode    VARCHAR(512),
  createdAt         TIMESTAMP NOT NULL ,
  updatedAt         TIMESTAMP NOT NULL
);

CREATE TABLE SSHKeys (
  fingerprint       VARCHAR(128) PRIMARY KEY ,
  key               VARCHAR(1024) NOT NULL UNIQUE ,
  description       VARCHAR(1024) NOT NULL ,
  accountUUID       VARCHAR(36) NOT NULL REFERENCES Accounts(uuid) ON DELETE CASCADE ,
  createdAt         TIMESTAMP NOT NULL ,
  updatedAt         TIMESTAMP NOT NULL
);

CREATE TABLE Clients (
  uuid              VARCHAR(36) PRIMARY KEY ,
  name              VARCHAR(512) NOT NULL UNIQUE ,      -- in oauth lingo this is the client_id
  secret            VARCHAR(512) ,
  redirectURIs      VARCHAR[] NOT NULL ,
  createdAt         TIMESTAMP NOT NULL ,
  updatedAt         TIMESTAMP NOT NULL
);

CREATE TABLE ClientScopeProvided (
  clientUUID        VARCHAR(36) NOT NULL REFERENCES Clients(uuid) ON DELETE CASCADE ,
  name              VARCHAR(512) NOT NULL UNIQUE ,
  description       VARCHAR(1024) NOT NULL
);

CREATE TABLE ClientApprovals (
  uuid              VARCHAR(36) PRIMARY KEY ,
  scope             VARCHAR[] NOT NULL ,
  clientUUID        VARCHAR(36) NOT NULL REFERENCES Clients(uuid) ON DELETE CASCADE ,
  accountUUID       VARCHAR(36) NOT NULL REFERENCES Accounts(uuid) ON DELETE CASCADE ,
  createdAt         TIMESTAMP NOT NULL ,
  updatedAt         TIMESTAMP NOT NULL ,
  UNIQUE (clientUUID, accountUUID)
);

CREATE TABLE GrantRequests (
  token             VARCHAR(512) PRIMARY KEY ,       -- the grant request id
  grantType         VARCHAR(10) NOT NULL CHECK (grantType = 'code' OR grantType = 'token'),
  state             VARCHAR(512) NOT NULL ,
  code              VARCHAR(512) ,
  scopeRequested    VARCHAR[] NOT NULL ,
  redirectURI       VARCHAR(512) NOT NULL ,
  clientUUID        VARCHAR(36) NOT NULL REFERENCES Clients(uuid) ON DELETE CASCADE ,
  accountUUID       VARCHAR(36) NULL REFERENCES Accounts(uuid) ON DELETE CASCADE ,
  createdAt         TIMESTAMP NOT NULL ,
  updatedAt         TIMESTAMP NOT NULL
);

CREATE TABLE RefreshTokens (
  token             VARCHAR(512) PRIMARY KEY ,
  scope             VARCHAR[] NOT NULL ,
  clientUUID        VARCHAR(36) NOT NULL REFERENCES Clients(uuid) ON DELETE CASCADE ,
  accountUUID       VARCHAR(36) NOT NULL REFERENCES Accounts(uuid) ON DELETE CASCADE ,
  createdAt         TIMESTAMP NOT NULL ,
  updatedAt         TIMESTAMP NOT NULL
);

CREATE TABLE AccessTokens (
  token             VARCHAR(512) PRIMARY KEY ,
  scope             VARCHAR[] NOT NULL ,
  expires           TIMESTAMP NOT NULL ,
  clientUUID        VARCHAR(36) NOT NULL REFERENCES Clients(uuid) ON DELETE CASCADE ,
  accountUUID       VARCHAR(36) NOT NULL REFERENCES Accounts(uuid) ON DELETE CASCADE ,
  createdAt         TIMESTAMP NOT NULL ,
  updatedAt         TIMESTAMP NOT NULL
);

CREATE TABLE Sessions (
  token             VARCHAR(512) PRIMARY KEY ,      -- the session id
  expires           TIMESTAMP NOT NULL ,
  accountUUID       VARCHAR(36) NOT NULL REFERENCES Accounts(uuid) ON DELETE CASCADE ,
  createdAt         TIMESTAMP NOT NULL ,
  updatedAt         TIMESTAMP NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back


DROP TABLE IF EXISTS Sessions CASCADE;
DROP TABLE IF EXISTS AccessTokens CASCADE;
DROP TABLE IF EXISTS RefreshTokens CASCADE;
DROP TABLE IF EXISTS ClientApprovals CASCADE;
DROP TABLE IF EXISTS GrantRequests CASCADE;
DROP TABLE IF EXISTS ClientScopeProvided CASCADE;
DROP TABLE IF EXISTS Clients CASCADE;
DROP TABLE IF EXISTS SSHKeys CASCADE;
DROP TABLE IF EXISTS Accounts CASCADE;
