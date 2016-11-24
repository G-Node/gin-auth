-- Copyright (c) 2016, German Neuroinformatics Node (G-Node),
--                           Michael Sonntag <dev@g-node.org>
-- All rights reserved.
--
-- Redistribution and use in source and binary forms, with or without
-- modification, are permitted under the terms of the BSD License. See
-- LICENSE file in the root of the Project.


-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

INSERT INTO Clients (uuid, name, secret, scopewhitelist, scopeblacklist, redirecturis, createdat, updatedat) VALUES
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'gin', 'secret', '{"account-create"}', '{"account-admin"}',
            '{"https://localhost:8081/login","http://localhost:8080/"}', now(), now()),
  ('177c56a4-57b4-4baf-a1a7-04f3d8e5b276', 'wb', 'secret', '{"account-read","repo-read"}',
            '{"account-admin"}', '{"https://localhost:8081/login"}', now(), now());

INSERT INTO ClientScopeProvided (clientuuid, name, description) VALUES
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'account-create', 'Create an account') ,
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'account-read', 'Read access to your account data') ,
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'account-write', 'Write access to your account data') ,
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'account-admin', 'Admin access to all account data') ,
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'repo-read', 'Read access to your repositories and repositories shared with you') ,
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'repo-write', 'Write access to your repositories and repositories you have write access to');


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DELETE FROM ClientScopeProvided;
DELETE FROM Clients;
