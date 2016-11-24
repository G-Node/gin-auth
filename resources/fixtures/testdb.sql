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

INSERT INTO SSHKeys (fingerprint, accountUUID, description, temporary, key, createdAt, updatedAt) VALUES
  ('A3tkBXFQWkjU6rzhkofY55G7tPR/Lmna4B+WEGVFXOQ', 'bf431618-f696-4dca-a95d-882618ce4ef9', 'Key from alice', false, 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDLtRNg1UHUf0k0ZlkfoYod9NoDPpOgx2AStEaEk/0bIKBqWJUNAZUfc6CHooKXTP3YakgqI7/BxV2pVgJIFBI4K9yGeLu76mwTpIZUTjEw/VoOaNP/vfV0LmXvQXstXMOZkmWt1rFaLsBpL9REP7XxteZYc2tjyVqy32GsVZHh6pPNes2q1Cf+awhkV/kXjup5AXwROLzqRvYBRs8oMPFDRZEGGax/Pp+r2GTB44M8YC0p7JAL3tLDDWsLVyygFA0OGhUffHmOGGf69uhh5JHhOjp49GEGftABdjnJznrVAI/71ySt0xWHJIOgMScsUGLYJtOZE/9KVrOQgZ1UAQML bar@foo', now(), now()),
  ('SpWwZAvumrAEqWQIUakTix/R2YR9aB795Px7vMKCqmw', 'bf431618-f696-4dca-a95d-882618ce4ef9', 'Other key from alice', false, 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC8NSbfR5nklp5TH/jtpE4vCUXl5UeifcoREvHgJflhVbRFoHVQrd3nMFw+IpVpAn6XeZdQOweY9lOq1I0Zv0qsysbVipe8Dsi8MI7EMM7lTLUgWXOtm0JXiHo7U/ymX5769Y/dV+KQ+yaGswaEYiqkUpMJ9sOWVXaa5Ly+wJLXClIVWiZgvY0c4O7UJIYsyEhLPWNsYQkT/DAFCZbb47dxfl2WFrdRkeO6Wh3IIbmm08+A0V9/AkdrmJ+ZoyU44LsCkzl5sQLs6oeLozkdwU+glYZEZ9SbGIlm5/oGrSENrAMF+mmSH+iXPpJ/9+NzIHw3rE5bJcUEl4kPd5OHidaf bar@foo', now(), now()),
  ('x9nS/Siw6cUy0qemb10V0dSK8YQYS2BKvV5KFowitUw', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'Key from bob', false, 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC857PNeLe38+Q/m9gbhq8fmjD0NuyMC9g2cTSz32+S9LoUUBqQhY0IvsbLLH+0uvlBEBVrLFN+D/bUgBlJc1I+8PZUtagGcjmdBwZgaePJY4ew1xGwN9yxiFI1ICyk6NN+7HEYrB81Bl1zuNs7vQU/cZGyAybSd5onPU772cy1+Ot3iYCfZm9dY613LgOP/I6yCVPlE+385qx6IoEPXuJxi8GneIn8vMOM0zk+kVOUmRHPcJfxsuhh3nt5n3bNiapp4kHX2MH1jEHGgnPco86Js8SSZVeh81oRAPLVL3TrlNPoRC41BnZfo3eXXsIORIzW8nKe3ij8OOuXjpIqYFOL bar@foo', now(), now()),
  ('XDKYPWTM9ffhH+MvRs/zrNVP7eoYLf5YG8/1BJrZCJw', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'Bobs old key', false, 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDGniH9xRg4qKuUKu4+m731Q3+EBG/u7VqeeYpAPgXp9UzeC/k6nHAzOWyS6K/ZTlu086rY1ZT0cIaminwNKsmkDeMRD0p3rUvmfCBhCD7BIR9eZCdpd3sKfzxMfrqPh3T9YFGDqA6muFyiqWLMF8+FqpCGItPaxCmo7DjAIu3yalCKMkApfSZ0mnb8ichwuez8uvocwHfA3Df946UgNTl1AtD3h1GNlt9xW6xaIYJIdFVZ6XoC/osejxudWppop69MzNPUAZdNOIxKDPkqXRiFIxjLL3Bu8fLZRUvFWGd3Vuf5nlB/fM+ckDHrrz7bZC2s8WecZZq644sIAJyHCuXD bar@foo', (now() - INTERVAL '1 day'), now()),
  ('LTPF+bl45+47oT1X+Yxy0oNH4P6xufQhNxGMjRvxP2A', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'Bobs old temporary key', true, 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDFvuAQeIhvyrf61heV+XeW4OBTmQpde1G29RSeuzG1UhGbLq/+ihiOYbH4ICL6LD8s5gSPSl50XBOSXZPObn0ZG6TjCwArGSpzEUtTh8nqmp583dDHdeBayfigqwGzZN7+GK8YGTqcwLXg/HpaFXthnS3eHAud9UqKZVtyTVcS5bRqs6BlHnSSxzcH8wZFgG2TtmQ3xJhUcSA7+XzA5CVrmgdD+Jr28kAkGFDmNz/7Smzk3O4wsEouwxyhxcAWxTBscVPUSAHvcFC8rHrFv25mWe/9KeIfhxzsq2rLQ/JXFF1XY3VKjSGC7kbi9oKE4/IBXnmh3VUgwCOxo6z7OkgN bar@foo', (now() - INTERVAL '1 day'), now()),
  ('dgU2JX3eCYur5xbKhFQ+jEACSurCwtRaG+Qn6SYq7lE', '51f5ac36-d332-4889-8023-6e033fcd8e17', 'Bobs new temporary key', true, 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDKHfQ67plrnKU5ua2JP6zTYZWiN23H26paJ4M/7r1/m9Ct8a3Oy5qK0LGmwj+nSInOX5U5AmQSnAfqnVcXG1QWP/GEvz7fxm+99ZU00P+Pti1AenmiK69qxvP7dMC3KJbwe6haEgVHNbDy3Uj1lW+cIH+FUkpuoLr5B6tCrXAUD+ZJrSAR3VlYMbAQ5W4ElU3Oh1gruacINCy3B83D3PVSumdgnPopYQdcFSVFv22fHGal4iw1T/M0Xfe7iQevLaEa/F+BwX8IAqNJb3mA+1JQbF0Vkfo+qxMtK3OUK0hZIYheH9H1OIl53RZ18jck0IWBgyo8chegSMoNtL3gzA6p bar@foo', now(), now());

INSERT INTO Clients (uuid, name, secret, scopeWhitelist, scopeBlacklist, redirectURIs, createdAt, updatedAt) VALUES
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'gin', 'secret', '{"account-create"}','{"account-admin"}','{"https://localhost:8081/login","http://localhost:8080/"}', now(), now()),
  ('177c56a4-57b4-4baf-a1a7-04f3d8e5b276', 'wb', 'secret', '{"account-read","repo-read"}','{"account-admin"}','{"https://localhost:8081/login"}', now(), now());

INSERT INTO ClientScopeProvided (clientuuid, name, description) VALUES
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'account-create', 'Create an account'),
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'account-read', 'Read access to your account data'),
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'account-write', 'Write access to your account data'),
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'account-admin', 'Admin access to all account data'),
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'repo-read', 'Read access to your repositories and repositories shared with you'),
  ('8b14d6bb-cae7-4163-bbd1-f3be46e43e31', 'repo-write', 'Write access to your repositories and repositories you have write access to');

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
  ('print', 'no-reply@g-node.org', '{"a@example.com"}', 'content2', now()),
  ('skip', 'no-reply@g-node.org', '{"b@example.com"}', 'content3', now());
