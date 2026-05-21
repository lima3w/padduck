# Licensing & Legal

---

## License

Padduck is licensed under the **GNU General Public License v3.0 (GPL-3.0)**.

[Full license text](https://github.com/lima3w/padduck/blob/main/LICENSE)

### Key Terms

**You are free to:**
- Use Padduck for any purpose, including commercially
- Study and modify the source code
- Distribute copies of Padduck
- Distribute your modified versions

**Under these conditions:**
- Source code must be made available for any distributed version (including hosted/SaaS)
- Modified versions must also be licensed under GPL-3.0
- You must state changes made to the code
- You cannot apply additional restrictions beyond GPL-3.0

### FAQ on GPL-3.0

**Can I use Padduck internally at my company?** Yes, with no restrictions.

**Can I offer Padduck as a hosted service?** Yes, but users must be able to obtain the source code of your version (GPL-3.0, AGPL-3.0 distinction aside — GPL-3.0 allows hosted use without source distribution requirements for the hosted service itself).

**Can I use Padduck code in my proprietary product?** Not without complying with GPL-3.0 (which typically means open-sourcing the combined work under GPL-3.0).

---

## Trademark Usage

The name **Padduck** and associated logos are the property of the project maintainers.

Guidelines:
- You may use "Padduck" to accurately describe the software (e.g., "We use Padduck for IPAM")
- You may not use "Padduck" to imply official endorsement without permission
- Modified versions should be clearly identified as modified (not the official Padduck release)

---

## Contribution License Agreement

By submitting a pull request or contribution to Padduck, you agree that:

1. Your contribution is your original work or you have the right to submit it
2. You grant the project a perpetual, irrevocable license to use your contribution under GPL-3.0
3. Your contribution does not include proprietary code belonging to your employer without authorization

No formal CLA signing is required — submitting a PR is sufficient acknowledgment.

---

## Privacy Policy

### Data Padduck Collects

When self-hosting Padduck, **you** are the data controller. Padduck stores:
- User accounts (username, email, hashed password)
- IP address assignments and metadata
- Audit logs (user actions and source IPs)
- Session data

### GDPR Features

Padduck includes built-in GDPR support:
- **Data export**: `GET /api/v1/auth/me/export` — download all your data as JSON
- **Deletion request**: `POST /api/v1/auth/me/deletion-request` — request account deletion
- **Privacy policy consent**: Track and update user consent via **My Settings → Privacy**
- **Admin GDPR delete**: `POST /api/v1/admin/users/:id/gdpr-delete` — permanently remove user data

### Data Ownership

All data stored in your self-hosted Padduck instance is under your control. Padduck does not transmit any user data to external services unless you explicitly configure integrations (webhooks, update checks, etc.).

See [Data Ownership Philosophy](Data-Ownership-Philosophy) for the project's philosophy on data ownership.

---

## Third-Party Licenses

Padduck uses open-source components. A Software Bill of Materials (SBOM) is generated on each CI run.

Key dependencies and their licenses:

| Component | License |
|-----------|---------|
| Go | BSD-3-Clause |
| Fiber v2 | MIT |
| pgx/v5 | MIT |
| React | MIT |
| Tailwind CSS | MIT |
| Vite | MIT |
| Recharts | MIT |
| Cytoscape | MIT |
| go-oidc | Apache-2.0 |
| crewjam/saml | BSD-2-Clause |
| go-ldap/ldap | MIT |
| pquerna/otp | Apache-2.0 |
| PostgreSQL | PostgreSQL License |

Full dependency list and SBOM available in the repository.

---

## Terms of Service

Padduck is self-hosted software. As a self-hoster, you set your own terms of service for your deployment. The project itself does not offer a hosted service with formal ToS.
