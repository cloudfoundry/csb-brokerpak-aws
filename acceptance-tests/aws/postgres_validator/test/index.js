const { Client } = require('pg')

// reject with error
// resolve with connection
function connectPostgres(params) {
    return new Promise((resolve) => {
        client = new Client(params)
        client.connect()
        resolve(client)
    })
}

function postgresCreateSchema(client) {
    console.log("postgresCreateSchema")
    return new Promise((resolve, reject) => {
      client.query("CREATE SCHEMA sampledb", (err, result) => {
            if (err) {
                reject(err)
            } else {
                resolve(client)
            }
        })
    })
}

function postgresCreateTable(client) {
    console.log("postgresCreateTable")
    return new Promise((resolve, reject) => {
      client.query("CREATE TABLE sampledb.customer (first_name character varying(45) NOT NULL)", (err, result) => {
            if (err) {
                reject(err)
            } else {
                resolve(client)
            }
        })
    })
}

module.exports = async function (credentials, runServer) {
    let content = ""
    return connectPostgres({
        host: credentials.hostname,
        user: credentials.username,
        password: credentials.password,
        port: credentials.port,
        database: credentials.name,
        ssl: credentials.use_tls
    }).then((client) => {
        return postgresCreateSchema(client)
    }).then((client) => {
        return postgresCreateTable(client)
    }).then(() => {
        runServer(content)
    }).catch((error) => {
        console.error(error)
        throw new Error("postgres test failed", error)
    })
}
