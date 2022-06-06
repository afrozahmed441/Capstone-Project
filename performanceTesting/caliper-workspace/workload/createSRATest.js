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
    for (let i = 1; i < 1000; i++) {
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

    const requestP = {
      contractId: this.roundArguments.contractId,
      contractFunction: "InitPatient",
      invokerIdentity: "Bob",
      invokerMspId: "Org1MSP",
      targetPeers: ["peer0.org1.example.com"],
      targetOrganization: ["org1"],
      // contractArguments: [this.pids[i]],
      readOnly: false,
      channel: "mychannel",
    };

    const requestD = {
      contractId: this.roundArguments.contractId,
      contractFunction: "InitDoctor",
      invokerIdentity: "Alice",
      invokerMspId: "Org2MSP",
      targetPeers: ["peer0.org2.example.com"],
      targetOrganizations: ["org2"],
      // contractArguments: [this.dids[i]],
      readOnly: false,
      channel: "mychannel",
    };

    await this.sutAdapter.sendRequests(requestP);
    await this.sutAdapter.sendRequests(requestD);
  }

  async submitTransaction() {
    /// create share request agreement for bob
    this.count++;
    const did = this.dids[this.count];
    const pid = this.pids[this.count];
    const request = {
      contractId: this.roundArguments.contractId,
      contractFunction: "CreateRequestAgreement",
      invokerIdentity: "Alice",
      invokerMspId: "Org2MSP",
      targetPeers: ["peer0.org2.example.com"],
      targetOrganization: ["org2"],
      contractArguments: [
        did,
        pid,
        "MEQCIHBH4F+6VtHOeYGS9GrIGzzMVtLa+WcVpQnPz3ArTIr/AiBTHXW3b7 jFkhG1D2kFwIIpHk98vUzR1511Hv9Me9ixjg==",
        "MEUCIQDDJs4tWf×G9qWg2HOuzoUrSmOeUaQDhA823DgQeJZDsQIgS5Szn+GIwq1WZnf3AbMpMsWbC6GBuJilLXB2s1rYFbo=",
        "Alice",
        "org2",
      ],
      readOnly: false,
      channel: "mychannel",
    };

    await this.sutAdapter.sendRequests(request);
  }

  async cleanupWorkloadModule() {
    for (let i = 1; i <= this.roundArguments.assets; i++) {
      /// delete the share request agreements for bob
      const request = {
        contractId: this.roundArguments.contractId,
        contractFunction: "DeleteRequestAgreement",
        invokerIdentity: "Bob",
        invokerMspId: "Org1MSP",
        targetPeers: ["peer0.org1.example.com"],
        targetOrganization: ["org1"],
        contractArguments: [this.pids[i]],
        readOnly: false,
        channel: "mychannel",
      };
      await this.sutAdapter.sendRequests(request);
    }
  }
}

function createWorkloadModule() {
  return new MyWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
