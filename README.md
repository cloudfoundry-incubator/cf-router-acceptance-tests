# Router Acceptance Tests 

This test suite exercises [Router](https://github.com/cloudfoundry-incubator/cf-routing-release).

## Running Acceptance tests

### Test setup

To run the Router Acceptance tests, you will need:
- a running router deployment
- make sure that all the dependencies of this module are installed in your GOPATH (alternatively if you have cloned the [cf-routing-release](https://github.com/cloudfoundry-incubator/cf-routing-release) make sure that your GOPATH is set to root directory of cf-routing-release)
- an environment variable `ROUTER_API_CONFIG` which points to a `.json` file that contains the router api endpoint

The following commands will create a config file `router_config.json` for a [bosh-lite](https://github.com/cloudfoundry/bosh-lite) installation and set the `ROUTER_API_CONFIG` environment variable to the path for this file. Edit `router_config.json` as appropriate for your environment.


```bash
cd ~/workspace/cf-routing-release/src/github.com/cloudfoundry-incubator/cf-tcp-router-acceptance-tests/
cat > router_config.json <<EOF
{
  "address": "10.244.8.2",
  "port": 9999,
  "diego_api_url": "http://10.244.16.6:8888"
}
EOF
export ROUTER_API_CONFIG=$PWD/router_config.json
```

### Running the tests

After correctly setting the `ROUTER_API_CONFIG` environment variable, the following command will run the tests:

```
    ginkgo -r -nodes=3 router
```
