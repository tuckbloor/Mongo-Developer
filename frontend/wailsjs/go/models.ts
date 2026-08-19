export namespace model {
	
	export class BackupRecord {
	    fileName: string;
	    createdAt: string;
	    scope: string;
	    database?: string;
	    databaseCount: number;
	    collectionCount: number;
	    documentCount: number;
	    indexCount: number;
	    sizeBytes: number;
	    sourceName?: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileName = source["fileName"];
	        this.createdAt = source["createdAt"];
	        this.scope = source["scope"];
	        this.database = source["database"];
	        this.databaseCount = source["databaseCount"];
	        this.collectionCount = source["collectionCount"];
	        this.documentCount = source["documentCount"];
	        this.indexCount = source["indexCount"];
	        this.sizeBytes = source["sizeBytes"];
	        this.sourceName = source["sourceName"];
	    }
	}
	export class BackupCreateResult {
	    backup: BackupRecord;
	    downloadPath?: string;
	    downloadError?: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupCreateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.backup = this.convertValues(source["backup"], BackupRecord);
	        this.downloadPath = source["downloadPath"];
	        this.downloadError = source["downloadError"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class CollectionInfo {
	    name: string;
	    type?: string;
	
	    static createFrom(source: any = {}) {
	        return new CollectionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	    }
	}
	export class CollectionStats {
	    documents: number;
	    indexes: number;
	    avgDocumentKb: number;
	    totalSizeMb: number;
	    storageSizeMb: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectionStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.documents = source["documents"];
	        this.indexes = source["indexes"];
	        this.avgDocumentKb = source["avgDocumentKb"];
	        this.totalSizeMb = source["totalSizeMb"];
	        this.storageSizeMb = source["storageSizeMb"];
	    }
	}
	export class ConnectionHealth {
	    connected: boolean;
	    pingMs: number;
	    mongoDbVersion?: string;
	    checkedAt: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionHealth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.pingMs = source["pingMs"];
	        this.mongoDbVersion = source["mongoDbVersion"];
	        this.checkedAt = source["checkedAt"];
	        this.error = source["error"];
	    }
	}
	export class ConnectionInput {
	    id?: string;
	    name: string;
	    uri: string;
	    host?: string;
	    port?: number;
	    username?: string;
	    password?: string;
	    authDb?: string;
	    tls: boolean;
	    readOnly: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.uri = source["uri"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.authDb = source["authDb"];
	        this.tls = source["tls"];
	        this.readOnly = source["readOnly"];
	    }
	}
	export class ConnectionProfile {
	    id: string;
	    name: string;
	    uri: string;
	    username?: string;
	    authDb?: string;
	    tls: boolean;
	    readOnly: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.uri = source["uri"];
	        this.username = source["username"];
	        this.authDb = source["authDb"];
	        this.tls = source["tls"];
	        this.readOnly = source["readOnly"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DatabaseInfo {
	    name: string;
	    size?: number;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	    }
	}
	export class Diagnostics {
	    version: string;
	    os: string;
	    arch: string;
	    goVersion: string;
	    configDirectory: string;
	    backupDirectory: string;
	    connectionName?: string;
	    connectionUri?: string;
	    connected: boolean;
	    pingMs: number;
	    mongoDbVersion?: string;
	    readOnly: boolean;
	    checkedAt: string;
	    connectionError?: string;
	
	    static createFrom(source: any = {}) {
	        return new Diagnostics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.goVersion = source["goVersion"];
	        this.configDirectory = source["configDirectory"];
	        this.backupDirectory = source["backupDirectory"];
	        this.connectionName = source["connectionName"];
	        this.connectionUri = source["connectionUri"];
	        this.connected = source["connected"];
	        this.pingMs = source["pingMs"];
	        this.mongoDbVersion = source["mongoDbVersion"];
	        this.readOnly = source["readOnly"];
	        this.checkedAt = source["checkedAt"];
	        this.connectionError = source["connectionError"];
	    }
	}
	export class DummyDataReport {
	    database: string;
	    collections: number;
	    documents: number;
	    indexes: number;
	    created: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DummyDataReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.database = source["database"];
	        this.collections = source["collections"];
	        this.documents = source["documents"];
	        this.indexes = source["indexes"];
	        this.created = source["created"];
	    }
	}
	export class ExplainResult {
	    mode: string;
	    database: string;
	    collection: string;
	    stage: string;
	    indexName?: string;
	    usesIndex: boolean;
	    collectionScan: boolean;
	    documentsExamined: number;
	    keysExamined: number;
	    returned: number;
	    executionMs: number;
	    suggestedIndex?: string;
	    summary: string;
	
	    static createFrom(source: any = {}) {
	        return new ExplainResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.database = source["database"];
	        this.collection = source["collection"];
	        this.stage = source["stage"];
	        this.indexName = source["indexName"];
	        this.usesIndex = source["usesIndex"];
	        this.collectionScan = source["collectionScan"];
	        this.documentsExamined = source["documentsExamined"];
	        this.keysExamined = source["keysExamined"];
	        this.returned = source["returned"];
	        this.executionMs = source["executionMs"];
	        this.suggestedIndex = source["suggestedIndex"];
	        this.summary = source["summary"];
	    }
	}
	export class FindRequest {
	    connectionId: string;
	    database: string;
	    collection: string;
	    filterJson: string;
	    sortJson?: string;
	    skip: number;
	    limit: number;
	
	    static createFrom(source: any = {}) {
	        return new FindRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.database = source["database"];
	        this.collection = source["collection"];
	        this.filterJson = source["filterJson"];
	        this.sortJson = source["sortJson"];
	        this.skip = source["skip"];
	        this.limit = source["limit"];
	    }
	}
	export class FindResult {
	    documents: string[];
	    total: number;
	    queryMs: number;
	
	    static createFrom(source: any = {}) {
	        return new FindResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.documents = source["documents"];
	        this.total = source["total"];
	        this.queryMs = source["queryMs"];
	    }
	}
	export class LaravelModelInfo {
	    className: string;
	    namespace?: string;
	    path: string;
	    collection: string;
	    connection?: string;
	    mongoModel: boolean;
	    usesLegacyField: boolean;
	    relationships?: string[];
	
	    static createFrom(source: any = {}) {
	        return new LaravelModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.className = source["className"];
	        this.namespace = source["namespace"];
	        this.path = source["path"];
	        this.collection = source["collection"];
	        this.connection = source["connection"];
	        this.mongoModel = source["mongoModel"];
	        this.usesLegacyField = source["usesLegacyField"];
	        this.relationships = source["relationships"];
	    }
	}
	export class LaravelProjectInfo {
	    path: string;
	    isLaravel: boolean;
	    laravelVersion?: string;
	    mongoPackageInstalled: boolean;
	    mongoPackageVersion?: string;
	    envExists: boolean;
	    mongoConfigured: boolean;
	    artisanExists: boolean;
	    models?: LaravelModelInfo[];
	
	    static createFrom(source: any = {}) {
	        return new LaravelProjectInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.isLaravel = source["isLaravel"];
	        this.laravelVersion = source["laravelVersion"];
	        this.mongoPackageInstalled = source["mongoPackageInstalled"];
	        this.mongoPackageVersion = source["mongoPackageVersion"];
	        this.envExists = source["envExists"];
	        this.mongoConfigured = source["mongoConfigured"];
	        this.artisanExists = source["artisanExists"];
	        this.models = this.convertValues(source["models"], LaravelModelInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RestoreReport {
	    fileName: string;
	    databaseCount: number;
	    collectionCount: number;
	    documentCount: number;
	    indexCount: number;
	    restoredAt: string;
	
	    static createFrom(source: any = {}) {
	        return new RestoreReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileName = source["fileName"];
	        this.databaseCount = source["databaseCount"];
	        this.collectionCount = source["collectionCount"];
	        this.documentCount = source["documentCount"];
	        this.indexCount = source["indexCount"];
	        this.restoredAt = source["restoredAt"];
	    }
	}

}

