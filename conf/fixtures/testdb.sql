-- Test fixtures to be used in tests
DELETE FROM Accounts;
INSERT INTO Accounts (uuid, login, email, firstName, lastName, password, activationCode, createdAt, updatedAt) VALUES
  ('bf431618-f696-4dca-a95d-882618ce4ef9', 'alice', 'aclic@foo.com', 'Alice', 'Goodchild', 'secret', NULL, '2015-01-01 01:00:00', '2015-02-02 01:00:00'),
  ('51f5ac36-d332-4889-8023-6e033fcd8e17', 'bob', 'bob@foo.com', 'Bob', 'Beaver', 'secret', NULL, '2015-01-01 01:00:00', '2015-02-02 01:00:00');