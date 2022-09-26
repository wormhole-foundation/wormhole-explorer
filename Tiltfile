# Disable telemetry by default
analytics_settings(False)

# Moar updates (default is 3)
update_settings(max_parallel_updates = 10)

# When running Tilt on a server, this can be used to set the public hostname Tilt runs on
# for service links in the UI to work.
config.define_string("webHost", False, "Public hostname for port forwards")

# Components
config.define_bool("mongo", False, "Enable mongo component")
config.define_bool("mongo-express", False, "Enable mongo-express component")
config.define_bool("fly", False, "Enable fly component")
config.define_bool("server", False, "Enable server component")
config.define_bool("web", False, "Enable web component")
config.define_bool("web_hot", False, "Enable how web component")

cfg = config.parse()
webHost = cfg.get("webHost", "localhost")
mongo = cfg.get("mongo", True)
mongoExpress = cfg.get("mongo-express", True)
fly = cfg.get("fly", True)
server = cfg.get("server", True)
web = cfg.get("web", True)
web_hot = cfg.get("web_hot", True)

if mongo:
    k8s_yaml("devnet/mongo-pvc.yaml")
    k8s_yaml("devnet/mongo-pv.yaml")

    k8s_yaml("devnet/mongo.yaml")

    k8s_resource(
        "mongo",
        port_forwards = [
            port_forward(27017, name = "Mongo [:27017]", host = webHost),
        ]
    )

if mongoExpress:
    k8s_yaml("devnet/mongo-express.yaml")
    
    k8s_resource(
        "mongo-express",
        port_forwards = [
            port_forward(8081, name = "Mongo Express [:8081]", host = webHost),
        ],
        resource_deps = ["mongo"]
    )

if fly:
    docker_build(
        ref = "fly",
        context = "fly",
        dockerfile = "fly/Dockerfile",
    )

    k8s_yaml("devnet/fly.yaml")
    
    k8s_resource(
        "fly",
        resource_deps = ["mongo"]
    )

if server:
    docker_build(
        ref = "server",
        context = "server",
        dockerfile = "server/Dockerfile",
    )

    k8s_yaml("devnet/server.yaml")
    
    k8s_resource(
        "server",
        port_forwards = [
            port_forward(4000, name = "Server [:4000]", host = webHost),
        ],
        resource_deps = ["mongo"]
    )

if web:
    entrypoint = "/app/node_modules/.bin/serve -s build -n"
    live_update = []
    if web_hot:
        entrypoint = "npm start"
        live_update = [
            sync("./web/public", "/app/public"),
            sync("./web/src", "/app/src"),
        ]

    docker_build(
        ref = "web",
        context = "web",
        dockerfile = "web/Dockerfile",
        entrypoint = entrypoint,
        live_update = live_update,
    )

    k8s_yaml("devnet/web.yaml")

    k8s_resource(
        "web",
        resource_deps = [],
        port_forwards = [
            port_forward(3000, name = "Web [:3000]", host = webHost),
        ]
    )
