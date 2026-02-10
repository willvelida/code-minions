# Assess Deployed Infrastructure

## Purpose

Query live cloud resources via MCP servers to map the deployed architecture and generate a confirmed data flow diagram. This is the first step before generating a threat model.

**Next step**: Once the architecture diagram is confirmed, proceed to `actions/generate-threat-model.md`.

---

## Flow

### Step 1: Determine Cloud Environment 🛑

Check what MCP servers are available for cloud infrastructure queries.

| Check | Action |
|-------|--------|
| MCP servers available | List available MCP servers that can query cloud resources |
| Cloud provider | Identify which cloud platform(s) are in use (Azure, AWS, GCP, etc.) |
| Multiple providers | If multiple MCP servers are available, note all of them |

If no cloud MCP server is detected:

**🛑 STOP**: Ask the user:
1. Which cloud provider is the infrastructure deployed to?
2. Is there an MCP server available for that provider?
3. Can they provide access or an alternative way to describe the infrastructure?

If an MCP server is detected but multiple are available:

**🛑 STOP**: Ask the user which cloud environment to assess, or whether to assess all of them.

**Success Criteria:**
- [ ] Cloud provider(s) identified
- [ ] MCP server(s) available and accessible
- [ ] User confirmed which environment(s) to assess

---

### Step 2: Query Deployed Resources

Use the available MCP server(s) to discover:

| Resource Type | What to Query | Why |
|--------------|---------------|-----|
| Compute | VMs, containers, functions, app services | Identifies running components |
| Networking | VNets, subnets, NSGs, firewalls, load balancers, DNS | Maps network topology and boundaries |
| Data stores | Databases, storage accounts, caches, queues | Identifies data at rest |
| Identity | IAM roles, service principals, managed identities | Maps access control |
| API gateways | API management, front doors, CDNs | Identifies external entry points |
| Secrets management | Key vaults, secret managers | Identifies secrets handling |
| Monitoring | Logging, alerting configurations | Identifies observability gaps |

Document each discovered resource with:

| Field | Details |
|-------|---------|
| Resource name | As deployed |
| Resource type | Service category |
| Region/location | Where it is deployed |
| Public exposure | Is it internet-facing? |
| Access controls | What identity/network controls are in place |
| Encryption | At rest and in transit |

**Success Criteria:**
- [ ] All resource types queried
- [ ] Deployed resources catalogued
- [ ] Public-facing resources flagged

---

### Step 3: Map Architecture

Combine discovered resources into an architecture map:

1. **Components**: Group related resources into logical components
2. **Data stores**: Identify all data persistence layers
3. **External entities**: Users, third-party integrations, internet-facing endpoints
4. **Data flows**: How data moves between components (infer from network rules and service connections)
5. **Trust boundaries**: Network boundaries, identity boundaries, public/private boundaries

**Success Criteria:**
- [ ] Resources grouped into logical components
- [ ] Data flows inferred from network and service configuration
- [ ] Trust boundaries mapped from network and identity controls

---

### Step 4: Generate Architecture Diagram 🛑

Produce a Mermaid data flow diagram following `standards/document-template.md`.

The diagram must include:
- All discovered components and data stores
- Data flows with labels
- Trust boundaries (network segments, public/private)
- Cloud-specific context (regions, resource groups where relevant)

**🛑 STOP**: Present the architecture diagram to the user. Ask:
1. Does this accurately represent your deployed infrastructure?
2. Are there resources or connections not captured by the MCP query?
3. Are the trust boundaries correct?

Do NOT proceed until the user confirms the diagram.

**Success Criteria:**
- [ ] Mermaid diagram generated
- [ ] All discovered resources represented
- [ ] User confirmed diagram accuracy

---

### Step 5: Hand Off to Threat Model Generation

Once the user has confirmed the architecture diagram, proceed to `actions/generate-threat-model.md` to:
1. Apply STRIDE analysis to the confirmed diagram
2. Generate a threat model diagram in Mermaid
3. Prioritise threats with the user
4. Produce the threat model document

**Success Criteria:**
- [ ] Architecture diagram confirmed by user
- [ ] Components, data flows, and trust boundaries documented
- [ ] User directed to `actions/generate-threat-model.md`
