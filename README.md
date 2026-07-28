# RedFish API Mock Server

A lightweight Go HTTP server that implements a mock RedFish API, providing endpoints for testing and development of RedFish-compatible applications.

## Features

- **RedFish 1.18.0 Specification Compliance** - Compatible with RedFish 5.0
- **Basic Authentication** - Default credentials: `admin` / `password`
- **Core Resource Collections** - Systems, Chassis, Managers, and UpdateService endpoints
- **Firmware Management** - Mock firmware inventory and update operations
- **Virtual Media OS Installation** - Stateful ISO mounting, one-time CD boot, and reset workflow
- **OEM Profiles** - Mock, Supermicro, Dell, and Cisco identities and resource conventions
- **OData Annotations** - Proper JSON responses with RedFish OData context

## Quick Start

### Prerequisites

- Go 1.23.6 or later

### Installation & Running

1. **Clone or download the project**

   ```bash
   git clone https://github.com/Metify-io/redfish_api_mock.git
   cd redfish_api_mock
   ```

2. **Install dependencies**

   ```bash
   go mod tidy
   ```

3. **Build and run**

   ```bash
   make run
   ```

   `make run` creates `config.json` from `config.json.default` when needed and
   preserves an existing `config.json`. Pass server arguments with `ARGS`:

   ```bash
   make run ARGS="-host 10.0.0.209 -port 8040"
   ```

   To only build the executable:

   ```bash
   make build
   ```

4. **Server starts on port 8080**

   ```
   Starting RedFish Mock Server on :8080
   Default credentials: admin / password
   ```

### Testing the API

Test the service root endpoint:

```bash
curl -u admin:password http://localhost:8080/redfish/v1/ | jq
```

## API Endpoints

### Service Root

- `GET /redfish/v1/` - RedFish service root with links to resource collections

### Computer Systems

- `GET /redfish/v1/Systems` - Collection of computer systems
- `GET /redfish/v1/Systems/{id}` - Individual computer system details
- `PATCH /redfish/v1/Systems/{id}` - Configure boot source override
- `POST /redfish/v1/Systems/{id}/Actions/ComputerSystem.Reset` - Reset the system

### Chassis

- `GET /redfish/v1/Chassis` - Collection of chassis
- `GET /redfish/v1/Chassis/{id}` - Individual chassis details

### Managers

- `GET /redfish/v1/Managers` - Collection of managers
- `GET /redfish/v1/Managers/{id}` - Individual manager details
- `GET /redfish/v1/Managers/{id}/VirtualMedia` - Virtual media collection
- `GET /redfish/v1/Managers/{id}/VirtualMedia/CD` - Virtual CD/DVD state
- `POST /redfish/v1/Managers/{id}/VirtualMedia/CD/Actions/VirtualMedia.InsertMedia` - Mount an ISO
- `POST /redfish/v1/Managers/{id}/VirtualMedia/CD/Actions/VirtualMedia.EjectMedia` - Unmount the ISO

### Update Service

- `GET /redfish/v1/UpdateService` - Update service information
- `GET /redfish/v1/UpdateService/FirmwareInventory` - Firmware inventory collection
- `GET /redfish/v1/UpdateService/FirmwareInventory/{id}` - Individual firmware component
- `POST /redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate` - Mock firmware update

## Authentication

Protected endpoints require HTTP Basic Authentication. The credentials are set
in the `authentication` section of `config.json`:

- **Username:** `admin`
- **Password:** `password`

## Mock Data

Set the top-level `oem` field in `config.json` to `mock`, `supermicro`, `dell`,
or `cisco`. Each profile supplies that OEM's default identity, OEM extension,
resource IDs, manager name, and virtual-media ID. The implementations live in
separate `oem_*.go` files so more vendor-specific behavior can be added without
changing the common endpoint handlers.

```json
{
  "oem": "dell"
}
```

The server reads the file at startup, so restart it after making changes. Other
configured fields override the selected profile's defaults; omitted fields
retain those defaults. Unknown field names and unsupported OEM names cause
startup to fail, which helps catch configuration typos.

For example, a profile can still be customized with:

```json
{
  "oem": "dell",
  "service_root": {
    "vendor": "Acme Corporation",
    "oem": {
      "Acme": {
        "@odata.type": "#AcmeExtensions.v1_0_0.ServiceRoot"
      }
    }
  },
  "system": {
    "manufacturer": "Acme Corporation",
    "model": "Acme Rack 42",
    "installation_status_oem_key": "Acme",
    "oem": {
      "Acme": {
        "AssetTag": "lab-server"
      }
    }
  }
}
```

The checked-in `config.json.default` supplies common mock hardware data and uses
the `mock` profile by default, preserving the original responses, including:

- **Systems:** Mock Server X1000 with 2 CPUs, 64GB RAM
- **Chassis:** 1U RackMount chassis
- **Managers:** BMC with firmware version 1.0.0
- **Firmware Inventory:** BIOS, BMC, and NIC components with version information

## Example Usage

### Get Service Root

```bash
curl http://localhost:8080/redfish/v1/ | jq
```

### List All Systems

```bash
curl -u admin:password http://localhost:8080/redfish/v1/Systems | jq
```

### Get System Details

```bash
curl -u admin:password http://localhost:8080/redfish/v1/Systems/1 | jq
```

### Perform Mock Firmware Update

```bash
curl -u admin:password -X POST \
  -H "Content-Type: application/json" \
  -d '{"ImageURI": "https://example.com/firmware.bin"}' \
  http://localhost:8080/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate
```

### Perform a Mock OS Installation

Mount an OS ISO:

```bash
curl -u admin:password -X POST \
  -H "Content-Type: application/json" \
  -d '{"Image":"https://example.com/os.iso","Inserted":true,"WriteProtected":true}' \
  http://localhost:8080/redfish/v1/Managers/1/VirtualMedia/CD/Actions/VirtualMedia.InsertMedia
```

The insert operation downloads the complete image (using `UserName` and `Password`
as HTTP Basic credentials when supplied) and verifies that it contains a valid
ISO-9660 primary volume descriptor. The media is only mounted after validation;
an invalid image returns `400 Bad Request`, while a download failure returns
`502 Bad Gateway`.

Configure a one-time boot from the virtual CD and restart the system:

```bash
curl -u admin:password -X PATCH \
  -H "Content-Type: application/json" \
  -d '{"Boot":{"BootSourceOverrideEnabled":"Once","BootSourceOverrideTarget":"Cd"}}' \
  http://localhost:8080/redfish/v1/Systems/1

curl -u admin:password -X POST \
  -H "Content-Type: application/json" \
  -d '{"ResetType":"ForceRestart"}' \
  http://localhost:8080/redfish/v1/Systems/1/Actions/ComputerSystem.Reset
```

The mock installation status is exposed at
`Oem.MockVendor.InstallationStatus` on `GET /redfish/v1/Systems/1`. It transitions
from `Ready` to `MediaMounted`, then `Installing` after the reset, and `Installed`
after two seconds. All state is in memory and resets when the server restarts.

## Development

### Project Structure

- `main.go` - Common Redfish resources, handlers, and server setup
- `oem.go` - OEM behavior interface and profile selection
- `oem_*.go` - Mock, Supermicro, Dell, and Cisco behavior profiles
- `go.mod` - Go module definition

### Building

```bash
make build
```

Create `dist/redfish_api_mock.tar.gz` containing the executable and the current
`config.json`:

```bash
make package
```

## RedFish Compliance

This mock server implements key RedFish concepts:

- **OData Context** - JSON-LD metadata for schema information
- **Resource Collections** - RESTful collections with member references
- **Proper HTTP Status Codes** - 200 OK, 202 Accepted, 401 Unauthorized, etc.
- **RedFish Headers** - OData-Version 4.0 header on all responses

## Use Cases

- **Development Testing** - Test RedFish client applications without hardware
- **CI/CD Integration** - Mock RedFish endpoints for automated testing
- **API Learning** - Explore RedFish API structure and responses
- **Prototyping** - Build RedFish-compatible tools before hardware deployment

## License

This project is available as open source under the terms specified by the repository license.
