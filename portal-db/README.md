# Portal DB
The Portal DB is the house for all core business logic for both PATH and the Portal. The Portal DB is a _highly opinionated_ implementation of a Postgres database that can be used to manage and administer both PATH and a UI on top of PATH.

## Interacting with the database
### `make` Targets:
- `make portal_db_up` creates the Portal DB with the base schema (`./init/001_schema.sql`) and runs the Portal DB on port `:5435`. 
- `make portal_db_down` stops running the local Portal DB. 
- `make portal_db_env` creates and inits the Database, and helps set up the local development environment.
- `make portal_db_clean` stops the local Portal DB and deletes the database and drops the schema. 

### `scripts`
Helper scripts exist to quickly populate the database with real data.
- `./scripts/hydrate-gateways.sh` - Retrieves all onchain data about a given `gateway` and populates the Portal DB
- `./scripts/hydrate-services.sh` - Retrieves all onchain data about a set of `services` and populates the Portal DB
- `./scripts/hydrate-applications.sh` - Retrieves all onchain data about a set of `applications` and populates the Portal DB

## Tools
### `psql` (REQUIRED)
- 🍎 Mac: `brew install postgresql`
- 🅰️ Arch: `pacman -S postgresql` 
- 🌀 Debian: `sudo apt-get install postgresql`

### `dbeaver` (RECOMMENDED)
It is _highly recommended_ to use a GUI Database Explorer in conjunction with the Portal DB. This allows a user to directly Create, Read, Update, and Delete (CRUD) database records in a GUI. We recommend `dbeaver`. 

:::tip ERD - Entity Relationship Diagrams
One reason we recommend `dbeaver` is its native functionality of creating an ERD - a visual tool for seeing how tables are interrelated in SQL. Once you have the Database running and `dbeaver` installed and configured, you can right-click on a schema and choose: "View Diagram" for an interactive ERD.
:::

**Install `dbeaver`**
- 🍎 Mac: `brew install --cask dbeaver-community`
- 🅰️ Arch: `paru -S dbeaver` 
- 🌀 Debian: 
``` 
sudo add-apt-repository ppa:serge-rider/dbeaver-ce
sudo apt-get update
sudo apt-get install dbeaver-ce   
```

**`dbeaver` Connection String Setup**
- Open `dbeaver`
- File > New Connection
- Select Postgres for the DB Driver
- Enter the connection details
:::tip Connection Details
You can get the required connection details from using `make portal_db_up`
:::
- Connect to the database and explore
