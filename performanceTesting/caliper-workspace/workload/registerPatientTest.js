"use strict";

const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
const { v4: uuidv4 } = require("uuid");

class MyWorkload extends WorkloadModuleBase {
  constructor() {
    super();
    this.pids = [];
  }

  initIDs(assets) {
    this.count = assets;
    for (let i = 0; i < assets; i++) {
      const pid = uuidv4() + "P";

      this.pids.push(pid);
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
  }

  async submitTransaction() {
    /// create share request agreement for bob
    this.count--;
    const pid = this.pids[this.count];

    const data = {
      firstName: "Bob",
      lastName: "Miller",
      age: 50,
      gender: "M",
      email: "bob@gmail.com",
      contactNumber: "1234567890",
      city: "LA",
      state: "CA",
      country: "US",
      type: "Patient",
    };

    const patientData = {
      pid,
      personalInfo: data,
    };

    const patientDataBytes = [...Buffer.from(JSON.stringify(patientData))];

    const transientData = {
      asset_data: patientDataBytes,
    };

    const request = {
      contractId: this.roundArguments.contractId,
      contractFunction: "RegisterPatient",
      invokerIdentity: "Bob",
      invokerMspId: "Org1MSP",
      transientMap: transientData,
      targetPeers: ["peer0.org1.example.com"],
      targetOrganization: ["org1"],
      readOnly: false,
      channel: "mychannel",
    };

    await this.sutAdapter.sendRequests(request);
  }

  async cleanupWorkloadModule() {}
}

function createWorkloadModule() {
  return new MyWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
