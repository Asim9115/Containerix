TOOLS = [
    {
        "name": "read_build_logs",
        "description": (
            "Read the docker build output to see the exact error and "
            "the build steps that preceded it. Always call this first"
        ),
        "input_scheme": {
            "type":"object",
            "properties":{
                "last_n_lines": {
                    "type": "integer",
                    "description": "How many lines to return. Default 80."
                }
            },
        },
    },
]