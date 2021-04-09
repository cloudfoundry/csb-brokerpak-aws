const restify = require('restify');
const vcapServices = require('vcap_services');
const testPostgres = require('./test')

function runServer(content) {
    const server = restify.createServer();
    server.get('/', (_, res, next) => {
        res.send(content)
        next()
    });

    server.listen(process.env.PORT || 8080, function () {
        console.log('%s listening at %s', server.name, server.url);
    });
}

async function runTest(credentials, testFunc) {
    try {
        await testFunc(credentials, runServer)
    } catch (e) {
        console.error(e)
    }
}

let credentials = vcapServices.findCredentials({ instance: { tags: 'postgres' } });

console.log("postgres credentials", credentials)
if (Object.keys(credentials).length > 0) {
    runTest(credentials, testPostgres).then(() => {
        console.log('Success')
    }).catch((err) => {
        console.error('Test failure:', err)
    })
} else {
    console.error('No credentials for tag: postgres')
}
