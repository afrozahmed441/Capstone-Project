"use strict";

const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
const { v4: uuidv4 } = require("uuid");

class MyWorkload extends WorkloadModuleBase {
  constructor() {
    super();
    this.pids = [];
    this.dids = [];
    this.count = 0;
  }

  /// set ids
  initIDs(assets) {
    this.count = assets + 1;
    for (let i = 1; i < 1502; i++) {
      const pid = i + "P";
      const did = i + "D";

      this.pids.push(pid);
      this.dids.push(did);
    }
  }

  async initializeWorkloadModule(
    workerIndex,
    totalWorkers,
    roundIndex,
    roundArguments,
    sutAdapter,
    sutContext
  ) {
    await super.initializeWorkloadModule(
      workerIndex,
      totalWorkers,
      roundIndex,
      roundArguments,
      sutAdapter,
      sutContext
    );

    this.initIDs(this.roundArguments.assets);
    console.log(
      `Worker ${this.workerIndex}: Creating test asset ${this.roundArguments.assets}`
    );
    /// check if needed dynamic id's
    const request = {
      contractId: this.roundArguments.contractId,
      contractFunction: "InitAccessRequestAgreements",
      invokerIdentity: "John",
      invokerMspId: "Org1MSP",
      targetPeers: ["peer0.org1.example.com"],
      targetOrganization: ["org1"],
      // contractArguments: [this.dids[i], this.pids[i], "John", "org1"],
      readOnly: false,
      channel: "mychannel",
    };

    await this.sutAdapter.sendRequests(request);
  }

  async submitTransaction() {
    /// create share request agreement for bob
    this.count++;
    const pid = this.pids[this.count];
    const request = {
      contractId: this.roundArguments.contractId,
      contractFunction: "ValidateDataAccessRequest",
      invokerIdentity: "Bob",
      invokerMspId: "Org1MSP",
      targetPeers: ["peer0.org1.example.com"],
      targetOrganization: ["org1"],
      contractArguments: [pid, "True"],
      readOnly: false,
      channel: "mychannel",
    };

    await this.sutAdapter.sendRequests(request);
  }

  async cleanupWorkloadModule() {
    // for (let i = 0; i < this.roundArguments.assets; i++) {
    //   /// delete the share request agreements for bob
    //   const request = {
    //     contractId: this.roundArguments.contractId,
    //     contractFunction: "DeleteDataAccessRequest",
    //     invokerIdentity: "Bob",
    //     invokerMspId: "Org1MSP",
    //     targetPeers: ["peer0.org1.example.com"],
    //     targetOrganization: ["org1"],
    //     contractArguments: [this.pids[i]],
    //     readOnly: false,
    //     channel: "mychannel",
    //   };
    //   await this.sutAdapter.sendRequests(request);
    // }
  }
}

function createWorkloadModule() {
  return new MyWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
