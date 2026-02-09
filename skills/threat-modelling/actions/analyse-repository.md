# Analyse Repository

## Purpose

Scan a codebase and infrastructure-as-code files to map the system architecture and generate a confirmed data flow diagram. This is the first step before generating a threat model.

**Next step**: Once the architecture diagram is confirmed, proceed to `actions/generate-threat-model.md`.

---

## Flow

### Step 1: Identify Repository Scope 🛑

Examine the repository to understand what is present:

| Look For | Examples | Why |
|----------|----------|-----|
| Application code | `src/`, `app/`, `lib/` | Identifies components and entry points |
| Infrastructure-as-code | Bicep, Terraform, CloudFormation, Pulumi files | Maps cloud resources and network topology |
| Configuration files | `docker-compose.yml`, Kubernetes manifests, `.env` files | Reveals services, ports, secrets handling |
| API definitions | OpenAPI specs, GraphQL schemas, gRPC protos | Identifies external interfaces and data contracts |
| Authentication/authorisation | Auth middleware, token handling, RBAC configs | Identifies trust boundaries |
| Data stores | Database migrations, connection strings, cache configs | Identifies data at rest |
| External dependencies | Package manifests, API client code | Identifies external trust boundaries |

**🛑 STOP**: Present a summary of what was found to the user. Ask if anything is missing or if there are components outside the repository that should be included.

**Success Criteria:**
- [ ] Repository contents catalogued
- [ ] IaC files identified (if present)
- [ ] User confirmed scope is complete

---

### Step 2: Map Architecture

Using the repository analysis, identify:

1. **Components**: Services, applications, functions, containers
2. **Data stores**: Databases, caches, queues, blob storage
3. **External entities**: Users, third-party APIs, external services
4. **Data flows**: How data moves between components
5. **Trust boundaries**: Where privilege levels change

Document each element with:

| Element | Details to Capture |
|---------|-------------------|
| Component | Name, technology, purpose |
| Data store | Type, what data it holds, encryption at rest |
| External entity | Who/what, authentication method |
| Data flow | Source, destination, protocol, data sensitivity |
| Trust boundary | What it separates, authentication mechanism |

**Success Criteria:**
- [ ] All components identified
- [ ] Data flows mapped
- [ ] Trust boundaries defined

---

### Step 3: Generate Architecture Diagram 🛑

Produce a Mermaid data flow diagram following `standards/document-template.md`.

The diagram must include:
- All identified components, data stores, and external entities
- Data flows with labels describing what data moves
- Trust boundaries clearly marked
- A legend if the diagram is complex

**🛑 STOP**: Present the architecture diagram to the user. Ask:
1. Does this accurately represent your system?
2. Are any components or data flows missing?
3. Are the trust boundaries correct?

Do NOT proceed to threat identification until the user confirms the diagram.

**Success Criteria:**
- [ ] Mermaid diagram generated
- [ ] All elements from Step 2 represented
- [ ] User confirmed diagram accuracy

---

### Step 4: Hand Off to Threat Model Generation

Once the user has confirmed the architecture diagram, proceed to `actions/generate-threat-model.md` to:
1. Apply STRIDE analysis to the confirmed diagram
2. Generate a threat model diagram in Mermaid
3. Prioritise threats with the user
4. Produce the threat model document

**Success Criteria:**
- [ ] Architecture diagram confirmed by user
- [ ] Components, data flows, and trust boundaries documented
- [ ] User directed to `actions/generate-threat-model.md`
