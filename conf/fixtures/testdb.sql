-- Test fixtures to be used in tests
DELETE FROM Accounts;
INSERT INTO Accounts (uuid, login, email, firstName, lastName, pwHash, activationCode, createdAt, updatedAt) VALUES
  ('bf431618-f696-4dca-a95d-882618ce4ef9', 'alice', 'aclic@foo.com', 'Alice', 'Goodchild', '', NULL, '2015-01-01 01:00:00', '2015-02-02 01:00:00'),
  ('51f5ac36-d332-4889-8023-6e033fcd8e17', 'bob', 'bob@foo.com', 'Bob', 'Beaver', '', NULL, '2015-01-01 01:00:00', '2015-02-02 01:00:00');
-- Set pw to 'testtest'
UPDATE Accounts SET pwHash = '$2a$10$kYB77ZPuIxon00ZPpk6APeAqi5J7aOPpqaPwS6riF40/RrfQ.EMlW';

DELETE FROM SSHKeys;
INSERT INTO SSHKeys (fingerprint, accountUUID, description, key, createdAt, updatedAt) VALUES
  ('SHA256:68a7N8YngrRrQF51SqLOONxILfaPa2A6ooW02Uiz+wM', 'bf431618-f696-4dca-a95d-882618ce4ef9', 'Key from alice', 'AAAAB3NzaC1yc2EAAAADAQABAAABAQDEyHeIbLkYIbGVgSBD6qWoW81NHlAEEZT+a/c/R/xbCSxaybBQXVGjc3zbbCEBiN5Y9UaxO1Cp/zYmUSbfgU5Vt6jydHiHTrJCfrLhnLnYW5SHdv4OeMtXVYKpimirBE9nSrA2TIbwrX6BurD7b09qQo+4S4BrOHEM9SJXhyjHM+ZtaKPaD4yove31KH2HUj7YL9XuCD050MH0ENBj0d686WdoFFBqlK0sKdLU1eOWCr/9zyUtEwm6BmQ1aCenwpQp4GYVrIRdPUqFxtd3KUoCco6wQfBb+rc23NEUzd1gdk4U5egBUjeld5CQUkhGTXV3Z0n89iZs9sZPs46ckrTl', now(), now()),
  ('SHA256:x9nS/Siw6cUy0qemb10V0dSK8YQYS2BKvV5KFowitUw', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'Key from bob', 'AAAAB3NzaC1yc2EAAAADAQABAAABAQC857PNeLe38+Q/m9gbhq8fmjD0NuyMC9g2cTSz32+S9LoUUBqQhY0IvsbLLH+0uvlBEBVrLFN+D/bUgBlJc1I+8PZUtagGcjmdBwZgaePJY4ew1xGwN9yxiFI1ICyk6NN+7HEYrB81Bl1zuNs7vQU/cZGyAybSd5onPU772cy1+Ot3iYCfZm9dY613LgOP/I6yCVPlE+385qx6IoEPXuJxi8GneIn8vMOM0zk+kVOUmRHPcJfxsuhh3nt5n3bNiapp4kHX2MH1jEHGgnPco86Js8SSZVeh81oRAPLVL3TrlNPoRC41BnZfo3eXXsIORIzW8nKe3ij8OOuXjpIqYFOL', now(), now());

DELETE FROM Clients;
INSERT INTO Clients (uuid, name, secret, scopeProvided, redirectURIs, createdAt, updatedAt) VALUES
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'gin', 'secret', '{"repo-read","repo-write"}', '{"https://localhost:8081/login"}', now(), now());

DELETE FROM ClientApprovals;
INSERT INTO ClientApprovals (uuid, scope, clientUUID, accountUUID, createdAt, updatedAt) VALUES
  ('31da7869-4593-4682-b9f2-5f47987aa5fc', '{"repo-read","repo-write"}', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now());

DELETE FROM GrantRequests;
INSERT INTO GrantRequests (token, grantType, state, code, scopeRequested, scopeApproved, redirectUri, clientUUID, accountUUID, createdAt, updatedAt) VALUES
  ('U7JIKKYI', 'code', 'OCQYDRYW', 'HGZQP6WE','{"repo-read","repo-write"}', '{"repo-read"}', 'https://localhost:8081/login', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now()),
  ('B4LIMIMB', 'code', '6Y4UTL24', 'C52KLSIZ','{"repo-read","repo-write"}', '{"repo-read"}', 'https://localhost:8081/login', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', '51f5ac36-d332-4889-8023-6e033fcd8e17', now(), now());

DELETE FROM Sessions;
INSERT INTO Sessions (token, expires, accountUUID, createdAt, updatedAt) VALUES
  ('DNM5RS3C', 'tomorrow', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now()),
  ('2MFZZUKI', 'yesterday', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'yesterday', 'yesterday');

DELETE FROM AccessTokens;
INSERT INTO AccessTokens (token, expires, scope, clientUUID, accountUUID, createdAt, updatedAt) VALUES
  ('3N7MP7M7', 'tomorrow', '{"repo-read","repo-write"}', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now()),
  ('LJ3W7ZFK', 'yesterday', '{"repo-read","repo-write"}', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'yesterday', 'yesterday');

DELETE FROM RefreshTokens;
INSERT INTO RefreshTokens (token, scope, clientUUID, accountUUID, createdAt, updatedAt) VALUES
  ('YYPTDSVZ', '{"repo-read","repo-write"}', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now()),
  ('4FKJVX3K', '{"repo-read","repo-write"}', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'yesterday', 'yesterday');
