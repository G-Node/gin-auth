-- Test fixtures to be used in tests
DELETE FROM EmailQueue;
DELETE FROM RefreshTokens;
DELETE FROM AccessTokens;
DELETE FROM Sessions;
DELETE FROM GrantRequests;
DELETE FROM ClientApprovals;
DELETE FROM ClientScopeProvided;
DELETE FROM Clients;
DELETE FROM SSHKeys;
DELETE FROM Accounts;

INSERT INTO Accounts (uuid, login, pwHash, email, isEmailPublic, title, firstName, lastName, institute, department, city, country, isAffiliationPublic, activationCode, createdAt, updatedAt) VALUES
  ('bf431618-f696-4dca-a95d-882618ce4ef9', 'alice', '', 'aclic@foo.com', FALSE, 'Dr.', 'Alice', 'Goodchild', 'LMU', 'Biology II', 'Munich', 'Germany', FALSE, NULL, '2015-01-01 01:00:00', '2015-02-02 01:00:00'),
  ('51f5ac36-d332-4889-8023-6e033fcd8e17', 'bob', '', 'bob@foo.com', FALSE, 'Mr.', 'Bob', 'Beaver', 'LMU', 'Biology II', 'Munich', 'Germany', TRUE, NULL, '2015-01-01 01:00:00', '2015-02-02 01:00:00'),
  ('03dcd573-1cce-4eb1-8b33-73860575da65', 'john', '', 'jj@example.com', FALSE, 'Mr.', 'John', 'Josephson', 'LMU', 'Biology II', 'Munich', 'Germany', TRUE, NULL, '2015-01-01 01:00:00', '2015-02-02 01:00:00');
-- Set pw to 'testtest'
UPDATE Accounts SET pwHash = '$2a$10$kYB77ZPuIxon00ZPpk6APeAqi5J7aOPpqaPwS6riF40/RrfQ.EMlW';

-- add account active and disabled testaccounts
INSERT INTO Accounts (uuid, login, pwhash, email, firstname, lastname, institute, department, city, country, activationcode, resetpwcode, isdisabled, createdat, updatedat) VALUES
  ('test0001-1234-6789-1234-678901234567', 'inact_log1', '', 'email1@example.com', 'fname', 'lname', 'inst', 'dep', 'cty', 'ctry', 'ac_a', NULL, FALSE, now(), now()),
  ('test0002-1234-6789-1234-678901234567', 'inact_log2', '', 'email2@example.com', 'fname', 'lname', 'inst', 'dep', 'cty', 'ctry', NULL, 'rc_a', FALSE, now(), now()),
  ('test0003-1234-6789-1234-678901234567', 'inact_log3', '', 'email3@example.com', 'fname', 'lname', 'inst', 'dep', 'cty', 'ctry', 'ac_c', 'rc_b', FALSE, now(), now()),
  ('test0004-1234-6789-1234-678901234567', 'inact_log4', '', 'email4@example.com', 'fname', 'lname', 'inst', 'dep', 'cty', 'ctry', NULL, NULL, TRUE, now(), now()),
  ('test0005-1234-6789-1234-678901234567', 'inact_log5', '', 'email5@example.com', 'fname', 'lname', 'inst', 'dep', 'cty', 'ctry', 'ac_b', NULL, TRUE, now(), now()),
  ('test0006-1234-6789-1234-678901234567', 'inact_log6', '', 'email6@example.com', 'fname', 'lname', 'inst', 'dep', 'cty', 'ctry', 'ac_d', 'rc_c', TRUE, now(), now());

INSERT INTO SSHKeys (fingerprint, accountUUID, description, key, createdAt, updatedAt) VALUES
  ('A3tkBXFQWkjU6rzhkofY55G7tPR_Lmna4B-WEGVFXOQ', 'bf431618-f696-4dca-a95d-882618ce4ef9', 'Key from alice', 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDLtRNg1UHUf0k0ZlkfoYod9NoDPpOgx2AStEaEk/0bIKBqWJUNAZUfc6CHooKXTP3YakgqI7/BxV2pVgJIFBI4K9yGeLu76mwTpIZUTjEw/VoOaNP/vfV0LmXvQXstXMOZkmWt1rFaLsBpL9REP7XxteZYc2tjyVqy32GsVZHh6pPNes2q1Cf+awhkV/kXjup5AXwROLzqRvYBRs8oMPFDRZEGGax/Pp+r2GTB44M8YC0p7JAL3tLDDWsLVyygFA0OGhUffHmOGGf69uhh5JHhOjp49GEGftABdjnJznrVAI/71ySt0xWHJIOgMScsUGLYJtOZE/9KVrOQgZ1UAQML bar@foo', now(), now()),
  ('x9nS_Siw6cUy0qemb10V0dSK8YQYS2BKvV5KFowitUw', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'Key from bob', 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC857PNeLe38+Q/m9gbhq8fmjD0NuyMC9g2cTSz32+S9LoUUBqQhY0IvsbLLH+0uvlBEBVrLFN+D/bUgBlJc1I+8PZUtagGcjmdBwZgaePJY4ew1xGwN9yxiFI1ICyk6NN+7HEYrB81Bl1zuNs7vQU/cZGyAybSd5onPU772cy1+Ot3iYCfZm9dY613LgOP/I6yCVPlE+385qx6IoEPXuJxi8GneIn8vMOM0zk+kVOUmRHPcJfxsuhh3nt5n3bNiapp4kHX2MH1jEHGgnPco86Js8SSZVeh81oRAPLVL3TrlNPoRC41BnZfo3eXXsIORIzW8nKe3ij8OOuXjpIqYFOL bar@foo', now(), now());

INSERT INTO Clients (uuid, name, secret, scopeWhitelist, scopeBlacklist, redirectURIs, createdAt, updatedAt) VALUES
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'gin', 'secret', '{}','{"account-admin"}','{"https://localhost:8081/login"}', now(), now()),
  ('177c56a4-57b4-4baf-a1a7-04f3d8e5b276', 'wb', 'secret', '{"account-read","repo-read"}','{"account-admin"}','{"https://localhost:8081/login"}', now(), now());

INSERT INTO ClientScopeProvided (clientuuid, name, description) VALUES
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'account-read', 'Read access to your account data'),
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'account-write', 'Write access to your account data'),
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'account-admin', 'Admin access to all account data'),
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'repo-read', 'Read access to your repositories and repositories shared with you'),
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'repo-write', 'Write acces to your repositories and repositories you have write access to');

INSERT INTO ClientApprovals (uuid, scope, clientUUID, accountUUID, createdAt, updatedAt) VALUES
  ('31da7869-4593-4682-b9f2-5f47987aa5fc', '{"repo-read","repo-write"}', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now()),
  ('ffde3769-cb45-43c1-8afd-4fb154ddf0b0', '{"repo-write","account-write"}', '177c56a4-57b4-4baf-a1a7-04f3d8e5b276', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now());

INSERT INTO GrantRequests (token, grantType, state, code, scopeRequested, redirectUri, clientUUID, accountUUID, createdAt, updatedAt) VALUES
  ('U7JIKKYI', 'code', 'OCQYDRYW', 'HGZQP6WE','{"repo-read","repo-write"}', 'https://localhost:8081/login', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now()),
  ('QH92T99D', 'code', 'HD58GHV9', NULL ,'{"account-read","repo-read"}', 'https://localhost:8081/login', '177c56a4-57b4-4baf-a1a7-04f3d8e5b276', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now()),
  ('B4LIMIMB', 'code', '6Y4UTL24', 'C52KLSIZ','{"repo-read","repo-write"}', 'https://localhost:8081/login', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', '51f5ac36-d332-4889-8023-6e033fcd8e17', now(), now()),
  ('AGTBAI3D', 'code', 'GBNAM23L', 'KWANG2G4','{"account-read"}', 'https://localhost:8081/login', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'yesterday', 'yesterday');

INSERT INTO Sessions (token, expires, accountUUID, createdAt, updatedAt) VALUES
  ('DNM5RS3C', 'tomorrow', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now()),
  ('4KDNO8T0', 'tomorrow', '51f5ac36-d332-4889-8023-6e033fcd8e17', now(), now()),
  ('2MFZZUKI', 'yesterday', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'yesterday', 'yesterday');

INSERT INTO AccessTokens (token, expires, scope, clientUUID, accountUUID, createdAt, updatedAt) VALUES
  ('3N7MP7M7', 'tomorrow', '{"account-read","account-write","repo-read","repo-write"}', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now()),
  ('LJ3W7ZFK', 'yesterday', '{"account-read","account-write","repo-read","repo-write"}', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'yesterday', 'yesterday'),
  ('KDEW57D4', 'tomorrow', '{"account-admin","repo-admin"}', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', '51f5ac36-d332-4889-8023-6e033fcd8e17', now(), now());

INSERT INTO RefreshTokens (token, scope, clientUUID, accountUUID, createdAt, updatedAt) VALUES
  ('YYPTDSVZ', '{"repo-read","repo-write"}', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'bf431618-f696-4dca-a95d-882618ce4ef9', now(), now()),
  ('4FKJVX3K', '{"repo-read","repo-write"}', '8b14d6bb-cae7-4163-bbd1-f3be46e43e31', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'yesterday', 'yesterday');

INSERT INTO EmailQueue (mode, sender, recipient, content, createdat) VALUES
  (NULL, 'no-reply@g-node.org', '{"a@b.com"}', 'content1', now()),
  ('print', 'no-reply@g-node.org', '{"a@b.com"}', 'content2', now()),
  ('skip', 'no-reply@g-node.org', '{"a@b.com"}', 'content3', now());
