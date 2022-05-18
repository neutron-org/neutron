/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { ProofOps } from "../tendermint/crypto/proof";
import { ResponseDeliverTx } from "../tendermint/abci/types";

export const protobufPackage =
  "lidofinance.interchainadapter.interchainqueries";

export interface MsgRegisterInterchainQuery {
  query_data: string;
  query_type: string;
  zone_id: string;
  connection_id: string;
  update_period: number;
  sender: string;
}

export interface MsgRegisterInterchainQueryResponse {
  id: number;
}

export interface MsgSubmitQueryResult {
  query_id: number;
  sender: string;
  result: QueryResult | undefined;
}

export interface QueryResult {
  kv_results: StorageValue[];
  txs: TxValue[];
  height: number;
}

export interface StorageValue {
  storage_prefix: string;
  key: Uint8Array;
  value: Uint8Array;
  Proof: ProofOps | undefined;
}

export interface TxValue {
  tx: ResponseDeliverTx | undefined;
  delivery_proof: MerkleProof | undefined;
  inclusion_proof: MerkleProof | undefined;
  height: number;
}

export interface MerkleProof {
  total: number;
  index: number;
  leaf_hash: Uint8Array;
  aunts: Uint8Array[];
}

export interface MsgSubmitQueryResultResponse {}

const baseMsgRegisterInterchainQuery: object = {
  query_data: "",
  query_type: "",
  zone_id: "",
  connection_id: "",
  update_period: 0,
  sender: "",
};

export const MsgRegisterInterchainQuery = {
  encode(
    message: MsgRegisterInterchainQuery,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.query_data !== "") {
      writer.uint32(10).string(message.query_data);
    }
    if (message.query_type !== "") {
      writer.uint32(18).string(message.query_type);
    }
    if (message.zone_id !== "") {
      writer.uint32(26).string(message.zone_id);
    }
    if (message.connection_id !== "") {
      writer.uint32(34).string(message.connection_id);
    }
    if (message.update_period !== 0) {
      writer.uint32(40).uint64(message.update_period);
    }
    if (message.sender !== "") {
      writer.uint32(50).string(message.sender);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgRegisterInterchainQuery {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgRegisterInterchainQuery,
    } as MsgRegisterInterchainQuery;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.query_data = reader.string();
          break;
        case 2:
          message.query_type = reader.string();
          break;
        case 3:
          message.zone_id = reader.string();
          break;
        case 4:
          message.connection_id = reader.string();
          break;
        case 5:
          message.update_period = longToNumber(reader.uint64() as Long);
          break;
        case 6:
          message.sender = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgRegisterInterchainQuery {
    const message = {
      ...baseMsgRegisterInterchainQuery,
    } as MsgRegisterInterchainQuery;
    if (object.query_data !== undefined && object.query_data !== null) {
      message.query_data = String(object.query_data);
    } else {
      message.query_data = "";
    }
    if (object.query_type !== undefined && object.query_type !== null) {
      message.query_type = String(object.query_type);
    } else {
      message.query_type = "";
    }
    if (object.zone_id !== undefined && object.zone_id !== null) {
      message.zone_id = String(object.zone_id);
    } else {
      message.zone_id = "";
    }
    if (object.connection_id !== undefined && object.connection_id !== null) {
      message.connection_id = String(object.connection_id);
    } else {
      message.connection_id = "";
    }
    if (object.update_period !== undefined && object.update_period !== null) {
      message.update_period = Number(object.update_period);
    } else {
      message.update_period = 0;
    }
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = String(object.sender);
    } else {
      message.sender = "";
    }
    return message;
  },

  toJSON(message: MsgRegisterInterchainQuery): unknown {
    const obj: any = {};
    message.query_data !== undefined && (obj.query_data = message.query_data);
    message.query_type !== undefined && (obj.query_type = message.query_type);
    message.zone_id !== undefined && (obj.zone_id = message.zone_id);
    message.connection_id !== undefined &&
      (obj.connection_id = message.connection_id);
    message.update_period !== undefined &&
      (obj.update_period = message.update_period);
    message.sender !== undefined && (obj.sender = message.sender);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgRegisterInterchainQuery>
  ): MsgRegisterInterchainQuery {
    const message = {
      ...baseMsgRegisterInterchainQuery,
    } as MsgRegisterInterchainQuery;
    if (object.query_data !== undefined && object.query_data !== null) {
      message.query_data = object.query_data;
    } else {
      message.query_data = "";
    }
    if (object.query_type !== undefined && object.query_type !== null) {
      message.query_type = object.query_type;
    } else {
      message.query_type = "";
    }
    if (object.zone_id !== undefined && object.zone_id !== null) {
      message.zone_id = object.zone_id;
    } else {
      message.zone_id = "";
    }
    if (object.connection_id !== undefined && object.connection_id !== null) {
      message.connection_id = object.connection_id;
    } else {
      message.connection_id = "";
    }
    if (object.update_period !== undefined && object.update_period !== null) {
      message.update_period = object.update_period;
    } else {
      message.update_period = 0;
    }
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = object.sender;
    } else {
      message.sender = "";
    }
    return message;
  },
};

const baseMsgRegisterInterchainQueryResponse: object = { id: 0 };

export const MsgRegisterInterchainQueryResponse = {
  encode(
    message: MsgRegisterInterchainQueryResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgRegisterInterchainQueryResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgRegisterInterchainQueryResponse,
    } as MsgRegisterInterchainQueryResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgRegisterInterchainQueryResponse {
    const message = {
      ...baseMsgRegisterInterchainQueryResponse,
    } as MsgRegisterInterchainQueryResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
    return message;
  },

  toJSON(message: MsgRegisterInterchainQueryResponse): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<MsgRegisterInterchainQueryResponse>
  ): MsgRegisterInterchainQueryResponse {
    const message = {
      ...baseMsgRegisterInterchainQueryResponse,
    } as MsgRegisterInterchainQueryResponse;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
    return message;
  },
};

const baseMsgSubmitQueryResult: object = { query_id: 0, sender: "" };

export const MsgSubmitQueryResult = {
  encode(
    message: MsgSubmitQueryResult,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.query_id !== 0) {
      writer.uint32(8).uint64(message.query_id);
    }
    if (message.sender !== "") {
      writer.uint32(18).string(message.sender);
    }
    if (message.result !== undefined) {
      QueryResult.encode(message.result, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MsgSubmitQueryResult {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgSubmitQueryResult } as MsgSubmitQueryResult;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.query_id = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.sender = reader.string();
          break;
        case 3:
          message.result = QueryResult.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgSubmitQueryResult {
    const message = { ...baseMsgSubmitQueryResult } as MsgSubmitQueryResult;
    if (object.query_id !== undefined && object.query_id !== null) {
      message.query_id = Number(object.query_id);
    } else {
      message.query_id = 0;
    }
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = String(object.sender);
    } else {
      message.sender = "";
    }
    if (object.result !== undefined && object.result !== null) {
      message.result = QueryResult.fromJSON(object.result);
    } else {
      message.result = undefined;
    }
    return message;
  },

  toJSON(message: MsgSubmitQueryResult): unknown {
    const obj: any = {};
    message.query_id !== undefined && (obj.query_id = message.query_id);
    message.sender !== undefined && (obj.sender = message.sender);
    message.result !== undefined &&
      (obj.result = message.result
        ? QueryResult.toJSON(message.result)
        : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<MsgSubmitQueryResult>): MsgSubmitQueryResult {
    const message = { ...baseMsgSubmitQueryResult } as MsgSubmitQueryResult;
    if (object.query_id !== undefined && object.query_id !== null) {
      message.query_id = object.query_id;
    } else {
      message.query_id = 0;
    }
    if (object.sender !== undefined && object.sender !== null) {
      message.sender = object.sender;
    } else {
      message.sender = "";
    }
    if (object.result !== undefined && object.result !== null) {
      message.result = QueryResult.fromPartial(object.result);
    } else {
      message.result = undefined;
    }
    return message;
  },
};

const baseQueryResult: object = { height: 0 };

export const QueryResult = {
  encode(message: QueryResult, writer: Writer = Writer.create()): Writer {
    for (const v of message.kv_results) {
      StorageValue.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    for (const v of message.txs) {
      TxValue.encode(v!, writer.uint32(18).fork()).ldelim();
    }
    if (message.height !== 0) {
      writer.uint32(24).uint64(message.height);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryResult {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryResult } as QueryResult;
    message.kv_results = [];
    message.txs = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.kv_results.push(StorageValue.decode(reader, reader.uint32()));
          break;
        case 2:
          message.txs.push(TxValue.decode(reader, reader.uint32()));
          break;
        case 3:
          message.height = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryResult {
    const message = { ...baseQueryResult } as QueryResult;
    message.kv_results = [];
    message.txs = [];
    if (object.kv_results !== undefined && object.kv_results !== null) {
      for (const e of object.kv_results) {
        message.kv_results.push(StorageValue.fromJSON(e));
      }
    }
    if (object.txs !== undefined && object.txs !== null) {
      for (const e of object.txs) {
        message.txs.push(TxValue.fromJSON(e));
      }
    }
    if (object.height !== undefined && object.height !== null) {
      message.height = Number(object.height);
    } else {
      message.height = 0;
    }
    return message;
  },

  toJSON(message: QueryResult): unknown {
    const obj: any = {};
    if (message.kv_results) {
      obj.kv_results = message.kv_results.map((e) =>
        e ? StorageValue.toJSON(e) : undefined
      );
    } else {
      obj.kv_results = [];
    }
    if (message.txs) {
      obj.txs = message.txs.map((e) => (e ? TxValue.toJSON(e) : undefined));
    } else {
      obj.txs = [];
    }
    message.height !== undefined && (obj.height = message.height);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryResult>): QueryResult {
    const message = { ...baseQueryResult } as QueryResult;
    message.kv_results = [];
    message.txs = [];
    if (object.kv_results !== undefined && object.kv_results !== null) {
      for (const e of object.kv_results) {
        message.kv_results.push(StorageValue.fromPartial(e));
      }
    }
    if (object.txs !== undefined && object.txs !== null) {
      for (const e of object.txs) {
        message.txs.push(TxValue.fromPartial(e));
      }
    }
    if (object.height !== undefined && object.height !== null) {
      message.height = object.height;
    } else {
      message.height = 0;
    }
    return message;
  },
};

const baseStorageValue: object = { storage_prefix: "" };

export const StorageValue = {
  encode(message: StorageValue, writer: Writer = Writer.create()): Writer {
    if (message.storage_prefix !== "") {
      writer.uint32(10).string(message.storage_prefix);
    }
    if (message.key.length !== 0) {
      writer.uint32(18).bytes(message.key);
    }
    if (message.value.length !== 0) {
      writer.uint32(26).bytes(message.value);
    }
    if (message.Proof !== undefined) {
      ProofOps.encode(message.Proof, writer.uint32(34).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): StorageValue {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseStorageValue } as StorageValue;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.storage_prefix = reader.string();
          break;
        case 2:
          message.key = reader.bytes();
          break;
        case 3:
          message.value = reader.bytes();
          break;
        case 4:
          message.Proof = ProofOps.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): StorageValue {
    const message = { ...baseStorageValue } as StorageValue;
    if (object.storage_prefix !== undefined && object.storage_prefix !== null) {
      message.storage_prefix = String(object.storage_prefix);
    } else {
      message.storage_prefix = "";
    }
    if (object.key !== undefined && object.key !== null) {
      message.key = bytesFromBase64(object.key);
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = bytesFromBase64(object.value);
    }
    if (object.Proof !== undefined && object.Proof !== null) {
      message.Proof = ProofOps.fromJSON(object.Proof);
    } else {
      message.Proof = undefined;
    }
    return message;
  },

  toJSON(message: StorageValue): unknown {
    const obj: any = {};
    message.storage_prefix !== undefined &&
      (obj.storage_prefix = message.storage_prefix);
    message.key !== undefined &&
      (obj.key = base64FromBytes(
        message.key !== undefined ? message.key : new Uint8Array()
      ));
    message.value !== undefined &&
      (obj.value = base64FromBytes(
        message.value !== undefined ? message.value : new Uint8Array()
      ));
    message.Proof !== undefined &&
      (obj.Proof = message.Proof ? ProofOps.toJSON(message.Proof) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<StorageValue>): StorageValue {
    const message = { ...baseStorageValue } as StorageValue;
    if (object.storage_prefix !== undefined && object.storage_prefix !== null) {
      message.storage_prefix = object.storage_prefix;
    } else {
      message.storage_prefix = "";
    }
    if (object.key !== undefined && object.key !== null) {
      message.key = object.key;
    } else {
      message.key = new Uint8Array();
    }
    if (object.value !== undefined && object.value !== null) {
      message.value = object.value;
    } else {
      message.value = new Uint8Array();
    }
    if (object.Proof !== undefined && object.Proof !== null) {
      message.Proof = ProofOps.fromPartial(object.Proof);
    } else {
      message.Proof = undefined;
    }
    return message;
  },
};

const baseTxValue: object = { height: 0 };

export const TxValue = {
  encode(message: TxValue, writer: Writer = Writer.create()): Writer {
    if (message.tx !== undefined) {
      ResponseDeliverTx.encode(message.tx, writer.uint32(10).fork()).ldelim();
    }
    if (message.delivery_proof !== undefined) {
      MerkleProof.encode(
        message.delivery_proof,
        writer.uint32(18).fork()
      ).ldelim();
    }
    if (message.inclusion_proof !== undefined) {
      MerkleProof.encode(
        message.inclusion_proof,
        writer.uint32(26).fork()
      ).ldelim();
    }
    if (message.height !== 0) {
      writer.uint32(32).uint64(message.height);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): TxValue {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseTxValue } as TxValue;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.tx = ResponseDeliverTx.decode(reader, reader.uint32());
          break;
        case 2:
          message.delivery_proof = MerkleProof.decode(reader, reader.uint32());
          break;
        case 3:
          message.inclusion_proof = MerkleProof.decode(reader, reader.uint32());
          break;
        case 4:
          message.height = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): TxValue {
    const message = { ...baseTxValue } as TxValue;
    if (object.tx !== undefined && object.tx !== null) {
      message.tx = ResponseDeliverTx.fromJSON(object.tx);
    } else {
      message.tx = undefined;
    }
    if (object.delivery_proof !== undefined && object.delivery_proof !== null) {
      message.delivery_proof = MerkleProof.fromJSON(object.delivery_proof);
    } else {
      message.delivery_proof = undefined;
    }
    if (
      object.inclusion_proof !== undefined &&
      object.inclusion_proof !== null
    ) {
      message.inclusion_proof = MerkleProof.fromJSON(object.inclusion_proof);
    } else {
      message.inclusion_proof = undefined;
    }
    if (object.height !== undefined && object.height !== null) {
      message.height = Number(object.height);
    } else {
      message.height = 0;
    }
    return message;
  },

  toJSON(message: TxValue): unknown {
    const obj: any = {};
    message.tx !== undefined &&
      (obj.tx = message.tx ? ResponseDeliverTx.toJSON(message.tx) : undefined);
    message.delivery_proof !== undefined &&
      (obj.delivery_proof = message.delivery_proof
        ? MerkleProof.toJSON(message.delivery_proof)
        : undefined);
    message.inclusion_proof !== undefined &&
      (obj.inclusion_proof = message.inclusion_proof
        ? MerkleProof.toJSON(message.inclusion_proof)
        : undefined);
    message.height !== undefined && (obj.height = message.height);
    return obj;
  },

  fromPartial(object: DeepPartial<TxValue>): TxValue {
    const message = { ...baseTxValue } as TxValue;
    if (object.tx !== undefined && object.tx !== null) {
      message.tx = ResponseDeliverTx.fromPartial(object.tx);
    } else {
      message.tx = undefined;
    }
    if (object.delivery_proof !== undefined && object.delivery_proof !== null) {
      message.delivery_proof = MerkleProof.fromPartial(object.delivery_proof);
    } else {
      message.delivery_proof = undefined;
    }
    if (
      object.inclusion_proof !== undefined &&
      object.inclusion_proof !== null
    ) {
      message.inclusion_proof = MerkleProof.fromPartial(object.inclusion_proof);
    } else {
      message.inclusion_proof = undefined;
    }
    if (object.height !== undefined && object.height !== null) {
      message.height = object.height;
    } else {
      message.height = 0;
    }
    return message;
  },
};

const baseMerkleProof: object = { total: 0, index: 0 };

export const MerkleProof = {
  encode(message: MerkleProof, writer: Writer = Writer.create()): Writer {
    if (message.total !== 0) {
      writer.uint32(8).int64(message.total);
    }
    if (message.index !== 0) {
      writer.uint32(16).int64(message.index);
    }
    if (message.leaf_hash.length !== 0) {
      writer.uint32(26).bytes(message.leaf_hash);
    }
    for (const v of message.aunts) {
      writer.uint32(34).bytes(v!);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): MerkleProof {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMerkleProof } as MerkleProof;
    message.aunts = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.total = longToNumber(reader.int64() as Long);
          break;
        case 2:
          message.index = longToNumber(reader.int64() as Long);
          break;
        case 3:
          message.leaf_hash = reader.bytes();
          break;
        case 4:
          message.aunts.push(reader.bytes());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MerkleProof {
    const message = { ...baseMerkleProof } as MerkleProof;
    message.aunts = [];
    if (object.total !== undefined && object.total !== null) {
      message.total = Number(object.total);
    } else {
      message.total = 0;
    }
    if (object.index !== undefined && object.index !== null) {
      message.index = Number(object.index);
    } else {
      message.index = 0;
    }
    if (object.leaf_hash !== undefined && object.leaf_hash !== null) {
      message.leaf_hash = bytesFromBase64(object.leaf_hash);
    }
    if (object.aunts !== undefined && object.aunts !== null) {
      for (const e of object.aunts) {
        message.aunts.push(bytesFromBase64(e));
      }
    }
    return message;
  },

  toJSON(message: MerkleProof): unknown {
    const obj: any = {};
    message.total !== undefined && (obj.total = message.total);
    message.index !== undefined && (obj.index = message.index);
    message.leaf_hash !== undefined &&
      (obj.leaf_hash = base64FromBytes(
        message.leaf_hash !== undefined ? message.leaf_hash : new Uint8Array()
      ));
    if (message.aunts) {
      obj.aunts = message.aunts.map((e) =>
        base64FromBytes(e !== undefined ? e : new Uint8Array())
      );
    } else {
      obj.aunts = [];
    }
    return obj;
  },

  fromPartial(object: DeepPartial<MerkleProof>): MerkleProof {
    const message = { ...baseMerkleProof } as MerkleProof;
    message.aunts = [];
    if (object.total !== undefined && object.total !== null) {
      message.total = object.total;
    } else {
      message.total = 0;
    }
    if (object.index !== undefined && object.index !== null) {
      message.index = object.index;
    } else {
      message.index = 0;
    }
    if (object.leaf_hash !== undefined && object.leaf_hash !== null) {
      message.leaf_hash = object.leaf_hash;
    } else {
      message.leaf_hash = new Uint8Array();
    }
    if (object.aunts !== undefined && object.aunts !== null) {
      for (const e of object.aunts) {
        message.aunts.push(e);
      }
    }
    return message;
  },
};

const baseMsgSubmitQueryResultResponse: object = {};

export const MsgSubmitQueryResultResponse = {
  encode(
    _: MsgSubmitQueryResultResponse,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): MsgSubmitQueryResultResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgSubmitQueryResultResponse,
    } as MsgSubmitQueryResultResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgSubmitQueryResultResponse {
    const message = {
      ...baseMsgSubmitQueryResultResponse,
    } as MsgSubmitQueryResultResponse;
    return message;
  },

  toJSON(_: MsgSubmitQueryResultResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<MsgSubmitQueryResultResponse>
  ): MsgSubmitQueryResultResponse {
    const message = {
      ...baseMsgSubmitQueryResultResponse,
    } as MsgSubmitQueryResultResponse;
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  RegisterInterchainQuery(
    request: MsgRegisterInterchainQuery
  ): Promise<MsgRegisterInterchainQueryResponse>;
  /** this line is used by starport scaffolding # proto/tx/rpc */
  SubmitQueryResult(
    request: MsgSubmitQueryResult
  ): Promise<MsgSubmitQueryResultResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  RegisterInterchainQuery(
    request: MsgRegisterInterchainQuery
  ): Promise<MsgRegisterInterchainQueryResponse> {
    const data = MsgRegisterInterchainQuery.encode(request).finish();
    const promise = this.rpc.request(
      "lidofinance.interchainadapter.interchainqueries.Msg",
      "RegisterInterchainQuery",
      data
    );
    return promise.then((data) =>
      MsgRegisterInterchainQueryResponse.decode(new Reader(data))
    );
  }

  SubmitQueryResult(
    request: MsgSubmitQueryResult
  ): Promise<MsgSubmitQueryResultResponse> {
    const data = MsgSubmitQueryResult.encode(request).finish();
    const promise = this.rpc.request(
      "lidofinance.interchainadapter.interchainqueries.Msg",
      "SubmitQueryResult",
      data
    );
    return promise.then((data) =>
      MsgSubmitQueryResultResponse.decode(new Reader(data))
    );
  }
}

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array
  ): Promise<Uint8Array>;
}

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

const atob: (b64: string) => string =
  globalThis.atob ||
  ((b64) => globalThis.Buffer.from(b64, "base64").toString("binary"));
function bytesFromBase64(b64: string): Uint8Array {
  const bin = atob(b64);
  const arr = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; ++i) {
    arr[i] = bin.charCodeAt(i);
  }
  return arr;
}

const btoa: (bin: string) => string =
  globalThis.btoa ||
  ((bin) => globalThis.Buffer.from(bin, "binary").toString("base64"));
function base64FromBytes(arr: Uint8Array): string {
  const bin: string[] = [];
  for (let i = 0; i < arr.byteLength; ++i) {
    bin.push(String.fromCharCode(arr[i]));
  }
  return btoa(bin.join(""));
}

type Builtin = Date | Function | Uint8Array | string | number | undefined;
export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

function longToNumber(long: Long): number {
  if (long.gt(Number.MAX_SAFE_INTEGER)) {
    throw new globalThis.Error("Value is larger than Number.MAX_SAFE_INTEGER");
  }
  return long.toNumber();
}

if (util.Long !== Long) {
  util.Long = Long as any;
  configure();
}
