'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class MyWorkloadModule extends WorkloadModuleBase {
    constructor() {
        super();
    }

    /**
     * Initializes the workload module.
     * This function is called only once by the Caliper framework before the start of the test.
     * It can be used to perform any necessary setup for the workload.
     * @param {BlockchainInterface} blockchain The blockchain interface for the current round.
     * @param {object} context The blockchain context for the current round.
     */
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        // Perform any necessary initialization here
        console.log('Initializing workload module');
        this.sutAdapter = sutAdapter;
        this.sutContext = sutContext;
        this.txIndex = 0;
    }

    /**
     * Performs the workload operation.
     * This function is called by the Caliper framework to execute the workload operation.
     * @return {Promise<object>} A promise that resolves when the operation is completed.
     */
    async submitTransaction() {
        // Perform the workload operation here
        const index = this.txIndex++;
        console.log(`Running workload operation ${index}`);

        // Call the smart contract function to initialize the ledger
        const request = {
            contractId: 'testgo', // Contract ID
            contractFunction: 'InitLedger', // Function name
            invokerIdentity: 'Admin@example.com', // Invoker identity
            contractArguments: [] // Arguments (none in this case)
        };

        await this.sutAdapter.sendRequests(request);
        return;
    }

    /**
     * Cleans up the workload module.
     * This function is called only once by the Caliper framework after the end of the test.
     * It can be used to perform any necessary cleanup for the workload.
     */
    async cleanupWorkloadModule() {
        // Perform any necessary cleanup here
        console.log('Cleaning up workload module');
    }
}

function createWorkloadModule() {
    return new MyWorkloadModule();
}

module.exports.createWorkloadModule = createWorkloadModule;
