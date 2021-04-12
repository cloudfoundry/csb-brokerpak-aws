const restify = require('restify');
const vcapServices = require('vcap_services');
const testPostgres = require('./test')

function runServer(content) {
    const server = restify.createServer();
    server.get('/', (_, res, next) => {
        res.send(content)
    });

    server.listen(process.env.PORT || 8080, function () {
        console.log('%s listening at %s', server.name, server.url);
    });
}

let credentials = vcapServices.findCredentials({ instance: { tags: 'postgres' } });

console.log("postgres credentials", credentials)
if (Object.keys(credentials).length > 0) {
      testPostgres(credentials, runServer).then(() => {
        console.log('Success')
      }).catch((err) => {
        console.error('Test failure:', e)
      })
} else {
    console.error('No credentials for tag: postgres')
}
