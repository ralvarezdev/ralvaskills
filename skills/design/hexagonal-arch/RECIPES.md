# Hexagonal Architecture Recipes

Reference implementations of the ports & adapters pattern from [SKILL.md](SKILL.md). Imports flow inward: adapter → domain, never the reverse.

## Go example

```go
// internal/users/user.go — domain (no imports outside stdlib)
package users

import "errors"

var ErrInactive = errors.New("user is inactive")

type User struct {
    ID    UserID
    Email Email
    State State
}

func (u *User) ChangeEmail(new Email) error {
    if u.State != Active {
        return ErrInactive
    }
    u.Email = new
    return nil
}

// internal/users/repo.go — port (interface declared by the consumer)
package users

import "context"

type Repo interface {
    FindByID(ctx context.Context, id UserID) (*User, error)
    Save(ctx context.Context, u *User) error
}

// internal/users/service.go — application service (depends on the port)
package users

type Service struct {
    repo Repo
}

func (s *Service) ChangeUserEmail(ctx context.Context, id UserID, email Email) error {
    u, err := s.repo.FindByID(ctx, id)
    if err != nil { return err }
    if err := u.ChangeEmail(email); err != nil { return err }
    return s.repo.Save(ctx, u)
}

// internal/users/postgres/repo.go — adapter (implements the port)
package postgres

import (
    "context"
    "_/myapp/internal/users"
    "github.com/jmoiron/sqlx"
)

type Repo struct{ db *sqlx.DB }

func (r *Repo) FindByID(ctx context.Context, id users.UserID) (*users.User, error) { /* ... */ }
func (r *Repo) Save(ctx context.Context, u *users.User) error                      { /* ... */ }
```

Imports flow: `postgres → users`, never the other direction.

## Python example

```python
# myapp/users/domain.py — domain (no framework imports)
from dataclasses import dataclass

class UserInactive(Exception): pass

@dataclass(slots=True)
class User:
    id: "UserID"
    email: "Email"
    state: "State"

    def change_email(self, new: "Email") -> None:
        if self.state is not State.ACTIVE:
            raise UserInactive
        self.email = new

# myapp/users/ports.py — port (Protocol)
from typing import Protocol

class UserRepo(Protocol):
    async def find_by_id(self, user_id: UserID) -> User: ...
    async def save(self, user: User) -> None: ...

# myapp/users/service.py — application service
class UserService:
    def __init__(self, repo: UserRepo) -> None:
        self._repo = repo

    async def change_user_email(self, user_id: UserID, email: Email) -> None:
        user = await self._repo.find_by_id(user_id)
        user.change_email(email)
        await self._repo.save(user)

# myapp/users/adapters/postgres.py — adapter
import psycopg
from myapp.users.domain import User
from myapp.users.ports import UserRepo

class PostgresUserRepo:
    def __init__(self, conn: psycopg.AsyncConnection) -> None:
        self._conn = conn

    async def find_by_id(self, user_id: UserID) -> User: ...
    async def save(self, user: User) -> None: ...
```

`adapters/postgres.py` imports `domain.py` and `ports.py`; the domain imports nothing in `adapters/`.
