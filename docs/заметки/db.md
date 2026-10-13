erDiagram
    User ||--o{ OrganizationMember : "has memberships"
    Organization ||--o{ OrganizationMember : "contains"
    Organization ||--o{ OrganizationInvite : "invites"
    User ||--o{ RefreshToken : "sessions"

    User {
        uuid id PK
        string email UK
        string password_hash
        string first_name
        string last_name
        boolean is_active
        boolean is_superuser
        timestamp created_at
        timestamp updated_at
    }

    Organization {
        uuid id PK
        string name
        string slug UK
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }

    OrganizationMember {
        uuid id PK
        uuid organization_id FK
        uuid user_id FK
        string role "owner | admin | manager | member"
        timestamp joined_at
    }

    RefreshToken {
        uuid id PK
        uuid user_id FK
        string token_hash UK
        string user_agent
        string ip_address
        timestamp expires_at
        boolean is_revoked
        timestamp created_at
    }
