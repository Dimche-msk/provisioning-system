# Technical Design: License Storage

## Overview
The Provisioning System uses a file-based license management system. The license is stored in a standardized JSON format, allowing for both online and offline verification and reporting.

## File Location
The license information is stored in a file named `license.key`. 
By default, this file is located in the configuration directory specified by the `--config-dir` flag (defaults to the current working directory).

Common path: `conf/license.key`

## Data Format
The `license.key` file is a UTF-8 encoded plain text file containing a JSON object.

### Fields
| Field | Type | Description |
| :--- | :--- | :--- |
| `tier` | string | License tier: `Free`, `Pro`, or `VIP`. |
| `customer_id` | string | Internal system identifier for the customer. |
| `issued_to` | string | Name of the person or organization to whom the license is issued. |
| `valid_from` | string (ISO8601) | The date and time when the license becomes valid. |
| `expiry` | string (ISO8601) | The date and time when the license expires. |
| `support_level` | string | The level of technical support included (e.g., `Community`, `VIP`, `Standard`). |
| `license_key` | string (UUID) | A unique identifier for the license instance. |

### Example
```json
{
  "tier": "Pro",
  "customer_id": "CUST-123456",
  "issued_to": "Acme Corp",
  "valid_from": "2026-01-01T00:00:00Z",
  "expiry": "2027-01-01T00:00:00Z",
  "support_level": "VIP",
  "license_key": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Management
- **Loading**: The backend `LicenseManager` loads this file on startup and during any "Reload Configuration" action.
- **Verification**: If the file is missing or invalid, the system defaults to the `Free` tier.
- **Reporting**: The license information is included in the `GetSystemStats` API response and embedded in the `system_info.json` file within generated support bundles for offline identification.
