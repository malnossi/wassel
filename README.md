# Wassel

Wassel is a secure, high-performance, cross-platform local file-sharing desktop application built with Go, Vue 3, and Wails. It operates entirely on your local area network (LAN) using zero-configuration discovery (mDNS) with an automated TCP subnet scanner fallback.

## 🚀 Key Features

*   **Zero-Configuration Discovery:** Automatically finds other devices running LocalShare on the same network using mDNS.
*   **Subnet Scanning Fallback:** Performs a bounded TCP subnet sweep if mDNS is restricted by your router or OS.
*   **Secure File Transfers:**
    *   **Download URL Host Validation:** Prevents URL injection or SSRF attacks by enforcing that the file download path aligns with the peer's actual TCP handshake IP.
    *   **Bounded Memory Consumption:** Employs size-limited stream parsing (`io.LimitReader`) for incoming JSON payloads to mitigate memory exhaustion vectors.
    *   **SHA-256 Integrity Verification:** Computes and verifies checksums for all transfers to detect data corruption or on-path tampering.
*   **Transfer Orchestration:** Seamlessly accept, decline, or cancel transfers in real-time from either side.
*   **Premium Modern UX:** Highly responsive interface constructed using a clean modular Vue 3 dashboard structure with HSL-tailored dark modes and radar discovery animations.

---

## 🏗️ Refactored Architecture

The codebase utilizes a clean, decoupled design separating the Wails IPC boundary from internal core packages:

```
localshare/
├── cmd/localshare/          # Bootstrapping (Wails options and main loop)
├── internal/
│   ├── app/                 # Wails app wrapper (thin IPC orchestration only)
│   ├── discovery/           # Peer store, mDNS services, and subnet scanner
│   ├── identity/            # Syrian cities device-name resolver
│   ├── platform/            # Cached network utilities (outbound IP detection)
│   └── transfer/            # Handshake listener, sender/receiver engines, and cancel managers
├── frontend/
│   ├── src/
│   │   ├── components/      # UI components (AppHeader, PeerPanel, TransferPanel, etc.)
│   │   ├── composables/     # Reactive state singletons (useDiscovery, useTransfers)
│   │   └── utils/           # formatBytes, formatSpeed UI formatting
│   └── wails.json           # Wails configuration
```

---

## 🛠️ Development & Building

### Prerequisites

Ensure you have the following installed:
*   [Go](https://golang.org/dl/) (version 1.22+)
*   [Node.js](https://nodejs.org/) & [pnpm](https://pnpm.io/)
*   [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### Live Development

Start the interactive live development server (backend builds on-the-fly and frontend supports hot-reload):
```bash
wails dev
```

### Production Build

Compile the final optimized production desktop bundle:
```bash
wails build
```

---

## 📝 License

This project is licensed under the MIT License.
