# RedFish API Mock Server

A lightweight Go HTTP server that implements a mock RedFish API, providing endpoints for testing and development of RedFish-compatible applications.

## Features

- **RedFish 1.18.0 Specification Compliance** - Compatible with RedFish 5.0
- **Basic Authentication** - Default credentials: `admin` / `password`
- **Core Resource Collections** - Systems, Chassis, Managers, and UpdateService endpoints
- **Firmware Management** - Mock firmware inventory and update operations
- **Virtual Media OS Installation** - Stateful ISO mounting, one-time CD boot, and reset workflow
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
   go run main.go
   ```

   Or build a binary:

   ```bash
   go build
   ./redfish_api_mock -host localhost -port 8080
   ```

   Mock identity and inventory values are loaded from `config.json`. To use a
   different file:

   ```bash
   ./redfish_api_mock -config /path/to/my-config.json
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

All endpoints require HTTP Basic Authentication:

- **Username:** `admin`
- **Password:** `password`

## Mock Data

Edit `config.json` to change the service root vendor and OEM extension, system,
chassis, manager, and firmware inventory returned by the API. The server reads
the file at startup, so restart it after making changes. Fields omitted from a
custom config retain their built-in defaults. Unknown field names cause startup
to fail, which helps catch configuration typos.

For example, an OEM-specific system can be configured with:

```json
{
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

The full checked-in `config.json` documents every supported setting and returns
the original mock data by default, including:

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

- `main.go` - Single file containing all server logic and data structures
- `go.mod` - Go module definition with gorilla/mux dependency

### Building

```bash
go build
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
