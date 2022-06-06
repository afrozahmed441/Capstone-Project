"use strict";

const { WorkloadModuleBase } = require("@hyperledger/caliper-core");
const { v4: uuidv4 } = require("uuid");

class MyWorkload extends WorkloadModuleBase {
  constructor() {
    super();
    this.dids = [];
  }

  initIDs(assets) {
    this.count = assets;
    for (let i = 0; i < assets; i++) {
      const did = uuidv4() + "D";

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
  }

  async submitTransaction() {
    /// create share request agreement for bob
    this.count--;
    const did = this.dids[this.count];

    const data = {
      firstName: "John",
      lastName: "Miles",
      age: 40,
      gender: "M",
      email: "john@gmail.com",
      contactNumber: "1234567890",
      city: "LA",
      state: "CA",
      country: "US",
      type: "Doctor",
    };

    const doctorData = {
      did,
      personalInfo: data,
      specialization: "Heart",
    };

    const doctorDataBytes = [...Buffer.from(JSON.stringify(doctorData))];

    const transientData = {
      asset_data: doctorDataBytes,
    };

    const request = {
      contractId: this.roundArguments.contractId,
      contractFunction: "RegisterDoctor",
      invokerIdentity: "John",
      transientMap: transientData,
      invokerMspId: "Org1MSP",
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
