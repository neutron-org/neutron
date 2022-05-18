/* eslint-disable */
import * as Long from "long";
import { util, configure, Writer, Reader } from "protobufjs/minimal";
import { Params } from "../interchainqueries/params";

export const protobufPackage =
  "lidofinance.interchainadapter.interchainqueries";

export interface RegisteredQuery {
  id: number;
  query_data: string;
  query_type: string;
  zone_id: string;
  connection_id: string;
  update_period: number;
  last_local_height: number;
  last_remote_height: number;
}

/** GenesisState defines the interchainadapter module's genesis state. */
export interface GenesisState {
  /** this line is used by starport scaffolding # genesis/proto/state */
  params: Params | undefined;
}

const baseRegisteredQuery: object = {
  id: 0,
  query_data: "",
  query_type: "",
  zone_id: "",
  connection_id: "",
  update_period: 0,
  last_local_height: 0,
  last_remote_height: 0,
};

export const RegisteredQuery = {
  encode(message: RegisteredQuery, writer: Writer = Writer.create()): Writer {
    if (message.id !== 0) {
      writer.uint32(8).uint64(message.id);
    }
    if (message.query_data !== "") {
      writer.uint32(18).string(message.query_data);
    }
    if (message.query_type !== "") {
      writer.uint32(26).string(message.query_type);
    }
    if (message.zone_id !== "") {
      writer.uint32(34).string(message.zone_id);
    }
    if (message.connection_id !== "") {
      writer.uint32(42).string(message.connection_id);
    }
    if (message.update_period !== 0) {
      writer.uint32(48).uint64(message.update_period);
    }
    if (message.last_local_height !== 0) {
      writer.uint32(56).uint64(message.last_local_height);
    }
    if (message.last_remote_height !== 0) {
      writer.uint32(64).uint64(message.last_remote_height);
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): RegisteredQuery {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseRegisteredQuery } as RegisteredQuery;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.id = longToNumber(reader.uint64() as Long);
          break;
        case 2:
          message.query_data = reader.string();
          break;
        case 3:
          message.query_type = reader.string();
          break;
        case 4:
          message.zone_id = reader.string();
          break;
        case 5:
          message.connection_id = reader.string();
          break;
        case 6:
          message.update_period = longToNumber(reader.uint64() as Long);
          break;
        case 7:
          message.last_local_height = longToNumber(reader.uint64() as Long);
          break;
        case 8:
          message.last_remote_height = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): RegisteredQuery {
    const message = { ...baseRegisteredQuery } as RegisteredQuery;
    if (object.id !== undefined && object.id !== null) {
      message.id = Number(object.id);
    } else {
      message.id = 0;
    }
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
    if (
      object.last_local_height !== undefined &&
      object.last_local_height !== null
    ) {
      message.last_local_height = Number(object.last_local_height);
    } else {
      message.last_local_height = 0;
    }
    if (
      object.last_remote_height !== undefined &&
      object.last_remote_height !== null
    ) {
      message.last_remote_height = Number(object.last_remote_height);
    } else {
      message.last_remote_height = 0;
    }
    return message;
  },

  toJSON(message: RegisteredQuery): unknown {
    const obj: any = {};
    message.id !== undefined && (obj.id = message.id);
    message.query_data !== undefined && (obj.query_data = message.query_data);
    message.query_type !== undefined && (obj.query_type = message.query_type);
    message.zone_id !== undefined && (obj.zone_id = message.zone_id);
    message.connection_id !== undefined &&
      (obj.connection_id = message.connection_id);
    message.update_period !== undefined &&
      (obj.update_period = message.update_period);
    message.last_local_height !== undefined &&
      (obj.last_local_height = message.last_local_height);
    message.last_remote_height !== undefined &&
      (obj.last_remote_height = message.last_remote_height);
    return obj;
  },

  fromPartial(object: DeepPartial<RegisteredQuery>): RegisteredQuery {
    const message = { ...baseRegisteredQuery } as RegisteredQuery;
    if (object.id !== undefined && object.id !== null) {
      message.id = object.id;
    } else {
      message.id = 0;
    }
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
    if (
      object.last_local_height !== undefined &&
      object.last_local_height !== null
    ) {
      message.last_local_height = object.last_local_height;
    } else {
      message.last_local_height = 0;
    }
    if (
      object.last_remote_height !== undefined &&
      object.last_remote_height !== null
    ) {
      message.last_remote_height = object.last_remote_height;
    } else {
      message.last_remote_height = 0;
    }
    return message;
  },
};

const baseGenesisState: object = {};

export const GenesisState = {
  encode(message: GenesisState, writer: Writer = Writer.create()): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): GenesisState {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseGenesisState } as GenesisState;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.params = Params.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },

  toJSON(message: GenesisState): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<GenesisState>): GenesisState {
    const message = { ...baseGenesisState } as GenesisState;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },
};

declare var self: any | undefined;
declare var window: any | undefined;
var globalThis: any = (() => {
  if (typeof globalThis !== "undefined") return globalThis;
  if (typeof self !== "undefined") return self;
  if (typeof window !== "undefined") return window;
  if (typeof global !== "undefined") return global;
  throw "Unable to locate global object";
})();

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
