{
	dtest: {
		debugPort: "",
		bootstrapTimeout: "",
		bootstrapReportInterval: "1m",
        nodePort: "",
        serviceID: "m3dbnode",
        dataDir: "",
        // loop through seeds to populate this list, but check if they're nil first
        seeds: [
			{
				namespace: "",
				localShardNum: "",
                retention: "",
                blockSize: "",
			}	
		], 
        // same thing here about instances
		instances: [
			{
				id: "",
				rack: "",
                zone: "",
                weight: 1,
				hostname: "",
			}
		],
		// end dtest
	},
    m3em: {
		agentPort: "",
		// do not include TLS for now, should be generated at the same time as k8 clusters 
		heartbeatPort: "",
        node: {
			heartbeat: {
				timeout: "4m",
		 		interval: "3s",	
			},
		},
        cluster: {
			replication: 3,
			numShards: 1024,
		    nodeConcurrency: 8,
			nodeOperationTimeout: "5m",	
		},
	},
	kv: {
		env: "",
		zone: "",
		service: "dtest",
		cacheDir: "/var/lib/m3kv",
		// loop thru dem clusters
		etcdClusters: [
			{
				zone: "",
				// loop thru the endpoints in the cluster
				endpoints: [
					"",
				],
			},
		]
	},
}
