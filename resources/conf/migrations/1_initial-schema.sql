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
  uuid                VARCHAR(36) PRIMARY KEY CHECK (char_length(uuid) = 36) ,
  login               VARCHAR(512) NOT NULL UNIQUE ,
  pwHash              VARCHAR(512) NOT NULL ,
  email               VARCHAR(512) NOT NULL UNIQUE ,
  isEmailPublic       BOOLEAN NOT NULL DEFAULT FALSE ,
  title               VARCHAR(512) ,
  firstName           VARCHAR(512) NOT NULL ,
  middleName          VARCHAR(512) ,
  lastName            VARCHAR(512) NOT NULL ,
  institute           VARCHAR(512) NOT NULL ,
  department          VARCHAR(512) NOT NULL ,
  city                VARCHAR(512) NOT NULL ,
  country             VARCHAR(512) NOT NULL ,
  isAffiliationPublic BOOLEAN NOT NULL DEFAULT FALSE ,
  activationCode      VARCHAR(512) UNIQUE,
  resetPWCode         VARCHAR(512) UNIQUE,
  isDisabled          BOOLEAN NOT NULL DEFAULT FALSE,
  createdAt           TIMESTAMP NOT NULL ,
  updatedAt           TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE VIEW ActiveAccounts AS
  SELECT * from Accounts
  WHERE NOT isDisabled AND activationCode IS NULL AND resetPWCode IS NULL;


CREATE TABLE SSHKeys (
  fingerprint       VARCHAR(128) PRIMARY KEY ,
  key               VARCHAR(1024) NOT NULL UNIQUE ,
  description       VARCHAR(1024) NOT NULL ,
  accountUUID       VARCHAR(36) NOT NULL REFERENCES Accounts(uuid) ,
  temporary         BOOLEAN NOT NULL DEFAULT FALSE ,
  createdAt         TIMESTAMP WITH TIME ZONE NOT NULL ,
  updatedAt         TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE Clients (
  uuid              VARCHAR(36) PRIMARY KEY CHECK (char_length(uuid) = 36),
  name              VARCHAR(512) NOT NULL UNIQUE CHECK (char_length(name) > 1),      -- in oauth lingo this is the client_id
  secret            VARCHAR(512) ,
  scopeWhitelist    VARCHAR[] NOT NULL ,
  scopeBlacklist    VARCHAR[] NOT NULL ,
  redirectURIs      VARCHAR[] NOT NULL ,
  createdAt         TIMESTAMP WITH TIME ZONE NOT NULL ,
  updatedAt         TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE ClientScopeProvided (
  clientUUID        VARCHAR(36) NOT NULL REFERENCES Clients(uuid) ON DELETE CASCADE ,
  name              VARCHAR(512) NOT NULL UNIQUE ,
  description       VARCHAR(1024) NOT NULL
);

CREATE TABLE ClientApprovals (
  uuid              VARCHAR(36) PRIMARY KEY CHECK (char_length(uuid) = 36) ,
  scope             VARCHAR[] NOT NULL ,
  clientUUID        VARCHAR(36) NOT NULL REFERENCES Clients(uuid) ON DELETE CASCADE ,
  accountUUID       VARCHAR(36) NOT NULL REFERENCES Accounts(uuid) ,
  createdAt         TIMESTAMP WITH TIME ZONE NOT NULL ,
  updatedAt         TIMESTAMP WITH TIME ZONE NOT NULL ,
  UNIQUE (clientUUID, accountUUID)
);

CREATE TABLE GrantRequests (
  token             VARCHAR(512) PRIMARY KEY ,       -- the grant request id
  grantType         VARCHAR(10) NOT NULL ,
  state             VARCHAR(512) NOT NULL ,
  code              VARCHAR(512) ,
  scopeRequested    VARCHAR[] NOT NULL ,
  redirectURI       VARCHAR(512) NOT NULL ,
  clientUUID        VARCHAR(36) NOT NULL REFERENCES Clients(uuid) ON DELETE CASCADE ,
  accountUUID       VARCHAR(36) NULL REFERENCES Accounts(uuid) ,
  createdAt         TIMESTAMP WITH TIME ZONE NOT NULL ,
  updatedAt         TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE RefreshTokens (
  token             VARCHAR(512) PRIMARY KEY ,
  scope             VARCHAR[] NOT NULL ,
  clientUUID        VARCHAR(36) NOT NULL REFERENCES Clients(uuid) ON DELETE CASCADE ,
  accountUUID       VARCHAR(36) NOT NULL REFERENCES Accounts(uuid) ,
  createdAt         TIMESTAMP WITH TIME ZONE NOT NULL ,
  updatedAt         TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE AccessTokens (
  token             VARCHAR(512) PRIMARY KEY ,
  scope             VARCHAR[] NOT NULL ,
  expires           TIMESTAMP WITH TIME ZONE NOT NULL ,
  clientUUID        VARCHAR(36) NOT NULL REFERENCES Clients(uuid) ON DELETE CASCADE ,
  accountUUID       VARCHAR(36) REFERENCES Accounts(uuid) ,
  createdAt         TIMESTAMP WITH TIME ZONE NOT NULL ,
  updatedAt         TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX ON AccessTokens (expires);

CREATE TABLE Sessions (
  token             VARCHAR(512) PRIMARY KEY ,      -- the session id
  expires           TIMESTAMP WITH TIME ZONE NOT NULL ,
  accountUUID       VARCHAR(36) NOT NULL REFERENCES Accounts(uuid) ,
  createdAt         TIMESTAMP WITH TIME ZONE NOT NULL ,
  updatedAt         TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX ON Sessions (expires);

CREATE TABLE EmailQueue (
  id        SERIAL PRIMARY KEY ,
  mode      VARCHAR(32) ,
  sender    VARCHAR(512) NOT NULL ,
  recipient VARCHAR[] NOT NULL ,
  content   VARCHAR(4096) NOT NULL ,
  createdAt TIMESTAMP WITH TIME ZONE NOT NULL
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP VIEW IF EXISTS ActiveAccounts;

DROP TABLE IF EXISTS EmailQueue CASCADE;
DROP TABLE IF EXISTS Sessions CASCADE;
DROP TABLE IF EXISTS AccessTokens CASCADE;
DROP TABLE IF EXISTS RefreshTokens CASCADE;
DROP TABLE IF EXISTS ClientApprovals CASCADE;
DROP TABLE IF EXISTS GrantRequests CASCADE;
DROP TABLE IF EXISTS ClientScopeProvided CASCADE;
DROP TABLE IF EXISTS Clients CASCADE;
DROP TABLE IF EXISTS SSHKeys CASCADE;
DROP TABLE IF EXISTS Accounts CASCADE;
