/* eslint-disable */
import { Reader, util, configure, Writer } from "protobufjs/minimal";
import * as Long from "long";
import { Params } from "../interchainqueries/params";
import { RegisteredQuery } from "../interchainqueries/genesis";
import { QueryResult } from "../interchainqueries/tx";

export const protobufPackage =
  "lidofinance.interchainadapter.interchainqueries";

/** QueryParamsRequest is request type for the Query/Params RPC method. */
export interface QueryParamsRequest {}

/** QueryParamsResponse is response type for the Query/Params RPC method. */
export interface QueryParamsResponse {
  /** params holds all the parameters of this module. */
  params: Params | undefined;
}

export interface QueryRegisteredQueriesRequest {}

export interface QueryRegisteredQueriesResponse {
  registered_queries: RegisteredQuery[];
}

export interface QueryRegisteredQueryResultRequest {
  query_id: number;
}

export interface QueryRegisteredQueryResultResponse {
  result: QueryResult | undefined;
}

const baseQueryParamsRequest: object = {};

export const QueryParamsRequest = {
  encode(_: QueryParamsRequest, writer: Writer = Writer.create()): Writer {
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
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

  fromJSON(_: any): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },

  toJSON(_: QueryParamsRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(_: DeepPartial<QueryParamsRequest>): QueryParamsRequest {
    const message = { ...baseQueryParamsRequest } as QueryParamsRequest;
    return message;
  },
};

const baseQueryParamsResponse: object = {};

export const QueryParamsResponse = {
  encode(
    message: QueryParamsResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.params !== undefined) {
      Params.encode(message.params, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: Reader | Uint8Array, length?: number): QueryParamsResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
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

  fromJSON(object: any): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromJSON(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },

  toJSON(message: QueryParamsResponse): unknown {
    const obj: any = {};
    message.params !== undefined &&
      (obj.params = message.params ? Params.toJSON(message.params) : undefined);
    return obj;
  },

  fromPartial(object: DeepPartial<QueryParamsResponse>): QueryParamsResponse {
    const message = { ...baseQueryParamsResponse } as QueryParamsResponse;
    if (object.params !== undefined && object.params !== null) {
      message.params = Params.fromPartial(object.params);
    } else {
      message.params = undefined;
    }
    return message;
  },
};

const baseQueryRegisteredQueriesRequest: object = {};

export const QueryRegisteredQueriesRequest = {
  encode(
    _: QueryRegisteredQueriesRequest,
    writer: Writer = Writer.create()
  ): Writer {
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryRegisteredQueriesRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryRegisteredQueriesRequest,
    } as QueryRegisteredQueriesRequest;
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

  fromJSON(_: any): QueryRegisteredQueriesRequest {
    const message = {
      ...baseQueryRegisteredQueriesRequest,
    } as QueryRegisteredQueriesRequest;
    return message;
  },

  toJSON(_: QueryRegisteredQueriesRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial(
    _: DeepPartial<QueryRegisteredQueriesRequest>
  ): QueryRegisteredQueriesRequest {
    const message = {
      ...baseQueryRegisteredQueriesRequest,
    } as QueryRegisteredQueriesRequest;
    return message;
  },
};

const baseQueryRegisteredQueriesResponse: object = {};

export const QueryRegisteredQueriesResponse = {
  encode(
    message: QueryRegisteredQueriesResponse,
    writer: Writer = Writer.create()
  ): Writer {
    for (const v of message.registered_queries) {
      RegisteredQuery.encode(v!, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryRegisteredQueriesResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryRegisteredQueriesResponse,
    } as QueryRegisteredQueriesResponse;
    message.registered_queries = [];
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.registered_queries.push(
            RegisteredQuery.decode(reader, reader.uint32())
          );
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryRegisteredQueriesResponse {
    const message = {
      ...baseQueryRegisteredQueriesResponse,
    } as QueryRegisteredQueriesResponse;
    message.registered_queries = [];
    if (
      object.registered_queries !== undefined &&
      object.registered_queries !== null
    ) {
      for (const e of object.registered_queries) {
        message.registered_queries.push(RegisteredQuery.fromJSON(e));
      }
    }
    return message;
  },

  toJSON(message: QueryRegisteredQueriesResponse): unknown {
    const obj: any = {};
    if (message.registered_queries) {
      obj.registered_queries = message.registered_queries.map((e) =>
        e ? RegisteredQuery.toJSON(e) : undefined
      );
    } else {
      obj.registered_queries = [];
    }
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryRegisteredQueriesResponse>
  ): QueryRegisteredQueriesResponse {
    const message = {
      ...baseQueryRegisteredQueriesResponse,
    } as QueryRegisteredQueriesResponse;
    message.registered_queries = [];
    if (
      object.registered_queries !== undefined &&
      object.registered_queries !== null
    ) {
      for (const e of object.registered_queries) {
        message.registered_queries.push(RegisteredQuery.fromPartial(e));
      }
    }
    return message;
  },
};

const baseQueryRegisteredQueryResultRequest: object = { query_id: 0 };

export const QueryRegisteredQueryResultRequest = {
  encode(
    message: QueryRegisteredQueryResultRequest,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.query_id !== 0) {
      writer.uint32(8).uint64(message.query_id);
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryRegisteredQueryResultRequest {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryRegisteredQueryResultRequest,
    } as QueryRegisteredQueryResultRequest;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.query_id = longToNumber(reader.uint64() as Long);
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryRegisteredQueryResultRequest {
    const message = {
      ...baseQueryRegisteredQueryResultRequest,
    } as QueryRegisteredQueryResultRequest;
    if (object.query_id !== undefined && object.query_id !== null) {
      message.query_id = Number(object.query_id);
    } else {
      message.query_id = 0;
    }
    return message;
  },

  toJSON(message: QueryRegisteredQueryResultRequest): unknown {
    const obj: any = {};
    message.query_id !== undefined && (obj.query_id = message.query_id);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryRegisteredQueryResultRequest>
  ): QueryRegisteredQueryResultRequest {
    const message = {
      ...baseQueryRegisteredQueryResultRequest,
    } as QueryRegisteredQueryResultRequest;
    if (object.query_id !== undefined && object.query_id !== null) {
      message.query_id = object.query_id;
    } else {
      message.query_id = 0;
    }
    return message;
  },
};

const baseQueryRegisteredQueryResultResponse: object = {};

export const QueryRegisteredQueryResultResponse = {
  encode(
    message: QueryRegisteredQueryResultResponse,
    writer: Writer = Writer.create()
  ): Writer {
    if (message.result !== undefined) {
      QueryResult.encode(message.result, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: Reader | Uint8Array,
    length?: number
  ): QueryRegisteredQueryResultResponse {
    const reader = input instanceof Uint8Array ? new Reader(input) : input;
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseQueryRegisteredQueryResultResponse,
    } as QueryRegisteredQueryResultResponse;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.result = QueryResult.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QueryRegisteredQueryResultResponse {
    const message = {
      ...baseQueryRegisteredQueryResultResponse,
    } as QueryRegisteredQueryResultResponse;
    if (object.result !== undefined && object.result !== null) {
      message.result = QueryResult.fromJSON(object.result);
    } else {
      message.result = undefined;
    }
    return message;
  },

  toJSON(message: QueryRegisteredQueryResultResponse): unknown {
    const obj: any = {};
    message.result !== undefined &&
      (obj.result = message.result
        ? QueryResult.toJSON(message.result)
        : undefined);
    return obj;
  },

  fromPartial(
    object: DeepPartial<QueryRegisteredQueryResultResponse>
  ): QueryRegisteredQueryResultResponse {
    const message = {
      ...baseQueryRegisteredQueryResultResponse,
    } as QueryRegisteredQueryResultResponse;
    if (object.result !== undefined && object.result !== null) {
      message.result = QueryResult.fromPartial(object.result);
    } else {
      message.result = undefined;
    }
    return message;
  },
};

/** Query defines the gRPC querier service. */
export interface Query {
  /** Parameters queries the parameters of the module. */
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse>;
  RegisteredQueries(
    request: QueryRegisteredQueriesRequest
  ): Promise<QueryRegisteredQueriesResponse>;
  QueryResult(
    request: QueryRegisteredQueryResultRequest
  ): Promise<QueryRegisteredQueryResultResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
  }
  Params(request: QueryParamsRequest): Promise<QueryParamsResponse> {
    const data = QueryParamsRequest.encode(request).finish();
    const promise = this.rpc.request(
      "lidofinance.interchainadapter.interchainqueries.Query",
      "Params",
      data
    );
    return promise.then((data) => QueryParamsResponse.decode(new Reader(data)));
  }

  RegisteredQueries(
    request: QueryRegisteredQueriesRequest
  ): Promise<QueryRegisteredQueriesResponse> {
    const data = QueryRegisteredQueriesRequest.encode(request).finish();
    const promise = this.rpc.request(
      "lidofinance.interchainadapter.interchainqueries.Query",
      "RegisteredQueries",
      data
    );
    return promise.then((data) =>
      QueryRegisteredQueriesResponse.decode(new Reader(data))
    );
  }

  QueryResult(
    request: QueryRegisteredQueryResultRequest
  ): Promise<QueryRegisteredQueryResultResponse> {
    const data = QueryRegisteredQueryResultRequest.encode(request).finish();
    const promise = this.rpc.request(
      "lidofinance.interchainadapter.interchainqueries.Query",
      "QueryResult",
      data
    );
    return promise.then((data) =>
      QueryRegisteredQueryResultResponse.decode(new Reader(data))
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
