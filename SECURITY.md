# Security Policy

## Supported Versions

Until Bearm reaches 1.0, security fixes are made on the latest development
branch. After stable releases begin, this document will list supported release
lines explicitly.

## Private Reporting

Report vulnerabilities through GitHub's private vulnerability reporting flow:

1. Open the Bearm repository's **Security** tab.
2. Select **Advisories**.
3. Select **Report a vulnerability**.

If private reporting is unavailable, contact the repository owner through the
private contact method shown on the GitHub profile. Do not open a public issue
for an unpatched vulnerability.

Include:

- affected Bearm version or commit;
- operating system and architecture;
- minimal reproduction using disposable paths;
- expected and observed safety behavior;
- impact and whether data loss or privilege boundaries are involved.

Do not send credentials, file contents, personal paths, or archives containing
private data. Maintainers will acknowledge receipt, investigate, coordinate a
fix, and publish an advisory when appropriate.

## Scope

Security-sensitive areas include:

- bypasses of hard or configured path protection;
- following a symlink destination;
- overwriting an existing trash item or metadata;
- unsafe trash-directory ownership or symlink handling;
- configuration injection or ownership bypass;
- journal corruption that causes unsafe restore or purge;
- permanent deletion without explicit native confirmation.

General compatibility differences without a safety impact belong in a normal
issue after checking the GNU and BSD compatibility documentation.
