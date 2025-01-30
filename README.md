# Rebate Management Backend
This is a simple backend to manage rebate, transaction and claims.
It uses sqllite as a db, this create a file called 'rebate.db' from whereever
server binary is deployed.

## Commands
* To build, run `make` on your terminal. This will create a server.out file
    Note:Requires `go build` to compile. Install GoLang compiler.
* To deploy locally run:`./server.out`, 
    This will host server on localhost:8080.do --help option to see available flags
## File Structure
All the code is currently under package main
* src/restapis:
    * main.go: Contain RunServer and initialization code
    * models.go: Contains all the ORM info
    * api\_handler.go: Contains all functionality
    * custom\_time.go: Implementation of CustomTime. It has to be implemented to
        translate request structue into ORMs
    * cache.go: Cache implementation
* cmd: Main folder for all executables
    * main.go: Executable entry code. Main function.
* Makefile
* README.md

# Endpoints
* POST   /rebate\_program: This api create a rebate program in system
* POST   /transaction: Creates a transaction
* GET    /rebate\_program/:rebate\_id: To view rebate on particular id
* GET    /calculate\_rebate/:transaction\_id: To calculate rebate on particular transaction
* POST   /claim\_rebate: To file a claim in the system
* GET    /reporting: This endpoint reports total amount of claims between input date range
* GET    /rebate\_claims/progress: This endpoint provides count of rebate claims in each category ie in 
            pending, approved and rejected.
## Data models
* Rebate Programs
    * program\_name: string
    * rebate\_percentage: real
    * start\_date: date in dd/mm/yyyy format
    * end\_date: date in dd/mm/yyyy format
         Only transaction between these dates will be eliible for this rebate.
    * eligibility\_criteria: string, For future use or documentation purposes
* Transaction
    * amount: real
    * transaction\_date: date in dd/mm/yyyy format
    * rebate\_program\_id: uint, id of associated rebate program
* Rebate Claims
    * transaction\_id: uint, Associated transaction id
    * claim\_amount: real, rebate amount calculated from program and transaction
    * claim\_status: string, one of 'pending','approved','rejected'

# Status
## Goals achieved
* All the requested endpoints have been implemented, with proper Business logic and error handlin
* Simple in-memory cache implemented
* Endpoint for realtime rebate claims monitor

## Standing problems
This program is bare bone and not functionally complete as of now.
Pending Requirements:
* CRUD apis for rebate, claims and transactions
* Configurable runtime options, like ip,port options, logging options etc
* Cache is only implemented for Rebate as a demo, it should also be implemented for
    transactions
* Code testing
* Containerization, for ease of deployment
# Further Improvements
1. ORM can be overhauled for better/extending use cases. Suggestions:
    * Rebate(R) and Transaction(T) should be kept independent
    * Claims(C) as relations b/w R and T
    * Single T can have multiplpe Rs
    * rebate amount make sense only at Cs not at Ts
    * Rs can have many checks as start date and end date, so a separate
        validation table with all validations aggregated.
    * Ts should have associated entities(company,firm etc) in db, these
        should also come into criteria for validating Rs.
2. Based on requirement, SQLLite can be upgraded to bette DBs. Options:distributed-SQL, columnar DB,
    GaphDB.
3. Endpoints need sanitization, based on user utility these endpoints needs to be reworked ie
    adding support for multi document, paginated responses
4. Security feature, user role and access restriction
5. Ellaborate Cache implementation

