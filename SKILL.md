# SKILL.md — LocalShare (Antigravity Architecture Specs)

This document serves as the formal specification, competency log, and engineering blueprint for **LocalShare**, a high-performance, local peer-to-peer file-sharing application built on top of the Antigravity (Go) runtime and Vue 3 + Element Plus.

---

## 🚀 Core Tech Stack

* **Backend Core:** Go (Golang)
* **Desktop Wrapper & Engine:** Antigravity Framework Runtime
* **Frontend Ecosystem:** Vue.js 3 (Composition API)
* **UI Components & System Layout:** Element Plus
* **Discovery Engine:** Multicast DNS (mDNS via standard link-local packages)
* **Data Distribution Protocol:** Native `net/http` sequential chunked streams

---

## 📐 System Architecture & Protocols

### 1. Zero-Configuration Discovery Layer (mDNS)
* **Mechanism:** Rather than standard raw UDP broadcast loops, LocalShare leverages Multicast DNS (`_localshare._tcp`) targeting the standard local domain multicast scope (`224.0.0.251:5353`).
* **Lifecycle Management:** Service announcements (`zeroconf.Register`) and passive browsing loops are bound directly to the global Antigravity lifecycle `context.Context`. If the backend runtime drops or hot-reloads, network resources are cleanly released.
* **Identity Mapping:** Devices are automatically cataloged using their native OS Hostnames to provide clear, human-readable representations within the interface.

### 2. File Transfer Engine (Sender-Hosted HTTP Streams)
* **Control/Handshake Channel:** Outbound transfer metadata is dispatched over a dedicated TCP stream bound to Port `9998`.
* **Data Channel:** * The **Sender** initializes an ephemeral, decoupled `net/http` server mapping a dynamic open port (`:0`) directly to the file resource pointer.
    * The **Receiver** catches the metadata block, triggers a user-facing authorization modal, and maps a single, high-speed `HTTP GET` request directly to the host's dynamic download URL.
* **Memory Footprint ($O(1)$ Optimization):** To guarantee system stability under heavy multi-gigabit workloads, data distribution avoids reading whole files into RAM. Files stream directly from disk to the network socket through standard chunked streaming.

### 3. IPC Interfacing & UI Update Throttling
* **Asynchronous Message Port Passing:** Data transfers between the Go core and the Vue webview use Antigravity’s structural RPC bindings (`app.Bind`) and custom event listeners (`app.Emit`).
* **Temporal Event Throttling:** During active gigabit local network streams, transfer updates are piped through a custom `ProgressWriter` that caps IPC event transmissions (`transfer:progress`) to a strict **150ms interval**. This preserves frontend responsiveness and keeps the UI fluid.

---

## 📊 Shared Data Models & IPC Schema

### 1. Peer Node Discovery Model
```json
{
  "hostname": "Mohameds-MacBook",
  "ip": "192.168.1.45"
}
```
### 2. Transactional Transfer Handshake Payload
```json
{
  "id": "a1b2c3d4",
  "filename": "archive_backup.tar.gz",
  "size": 524288000,
  "download_url": "[http://192.168.1.45:51234/download](http://192.168.1.45:51234/download)"
}
```
### 3. Throttled UI Progress Packet
```json
{
  "filename": "archive_backup.tar.gz",
  "percentage": 42.5,
  "current": 222822400,
  "total": 524288000
}
```
## 🌐 Application Layout (Antigravity Requirements)
To compile seamlessly using Antigravity, the repository enforces a clear separation between the compiled system binaries and the static web view layers:

```bash
LocalShare/
├── cmd/
│   └── localshare/
│       └── main.go         # Application entry point & Antigravity runtime boots
├── pkg/
│   ├── discovery/
│   │   └── service.go      # mDNS broadcasting and collection routines
│   └── transfer/
│       ├── engine.go       # Handshake mechanics & temporary HTTP host servers
│       └── progress.go     # Throttled progress streaming writers
├── ui/                     # Production compiled frontend directory (static distribution)
│   ├── index.html          # Entry document
│   └── assets/             # Bundled Vue 3 + Element Plus CSS/JS packages
├── antigravity.json        # Declared window bounds & static asset directory pathing
└── SKILL.md                # Structural blueprint log
```
## 🗺️ Milestone Tracking
[x] Milestone 1: Dynamic Peer Discovery

Configured high-reliability mDNS service registration (_localshare._tcp).

Linked discovery threads to standard Go context lifecycles.

[x] Milestone 2: Vue 3 + Element Plus Frontend Setup

Installed and styled clean dashboard views using Element Plus components.

Implemented isolated reactive composables (useNetwork.js) for system state management.

[x] Milestone 3: HTTP Streaming File Transfer

Built ephemeral sender-hosted HTTP file servers allocating random open system ports.

Authored low-level network handshake protocols running over raw TCP loops.

[x] Milestone 4: Throttled Progress Monitoring

Created a custom ProgressWriter layer with a 150ms update limiter to protect frontend performance.

Integrated native Element Plus <el-progress> bars for real-time tracking.