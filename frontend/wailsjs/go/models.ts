export namespace errors {
	
	export class ErrorInfo {
	    code: string;
	    title: string;
	    description: string;
	    suggestion: string;
	    retryable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ErrorInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.suggestion = source["suggestion"];
	        this.retryable = source["retryable"];
	    }
	}

}

export namespace main {
	
	export class ConnectionHistory {
	    // Go type: time
	    plexLastConnected?: any;
	    // Go type: time
	    discordLastConnected?: any;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionHistory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.plexLastConnected = this.convertValues(source["plexLastConnected"], null);
	        this.discordLastConnected = this.convertValues(source["discordLastConnected"], null);
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
	export class PlexConnectionStatus {
	    connected: boolean;
	    polling: boolean;
	    inErrorState: boolean;
	    serverUrl: string;
	    userId: string;
	    userName: string;
	
	    static createFrom(source: any = {}) {
	        return new PlexConnectionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.polling = source["polling"];
	        this.inErrorState = source["inErrorState"];
	        this.serverUrl = source["serverUrl"];
	        this.userId = source["userId"];
	        this.userName = source["userName"];
	    }
	}
	export class ResourceStats {
	    memoryAllocMB: number;
	    memoryTotalMB: number;
	    goroutineCount: number;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new ResourceStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.memoryAllocMB = source["memoryAllocMB"];
	        this.memoryTotalMB = source["memoryTotalMB"];
	        this.goroutineCount = source["goroutineCount"];
	        this.timestamp = source["timestamp"];
	    }
	}

}

export namespace plex {
	
	export class MusicSession {
	    sessionKey: string;
	    userId: string;
	    userName: string;
	    type: string;
	    state: string;
	    playerName: string;
	    track: string;
	    artist: string;
	    album: string;
	    thumb: string;
	    thumbUrl: string;
	    duration: number;
	    viewOffset: number;
	
	    static createFrom(source: any = {}) {
	        return new MusicSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionKey = source["sessionKey"];
	        this.userId = source["userId"];
	        this.userName = source["userName"];
	        this.type = source["type"];
	        this.state = source["state"];
	        this.playerName = source["playerName"];
	        this.track = source["track"];
	        this.artist = source["artist"];
	        this.album = source["album"];
	        this.thumb = source["thumb"];
	        this.thumbUrl = source["thumbUrl"];
	        this.duration = source["duration"];
	        this.viewOffset = source["viewOffset"];
	    }
	}
	export class PlexUser {
	    id: string;
	    name: string;
	    thumb: string;
	
	    static createFrom(source: any = {}) {
	        return new PlexUser(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.thumb = source["thumb"];
	    }
	}
	export class Server {
	    id: string;
	    name: string;
	    address: string;
	    port: string;
	    isLocal: boolean;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new Server(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.address = source["address"];
	        this.port = source["port"];
	        this.isLocal = source["isLocal"];
	        this.version = source["version"];
	    }
	}
	export class ValidationResult {
	    success: boolean;
	    serverName: string;
	    serverVersion: string;
	    libraryCount: number;
	    machineIdentifier: string;
	
	    static createFrom(source: any = {}) {
	        return new ValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.serverName = source["serverName"];
	        this.serverVersion = source["serverVersion"];
	        this.libraryCount = source["libraryCount"];
	        this.machineIdentifier = source["machineIdentifier"];
	    }
	}

}

export namespace retry {
	
	export class RetryState {
	    attemptNumber: number;
	    nextRetryIn: number;
	    // Go type: time
	    nextRetryAt: any;
	    lastError?: string;
	    lastErrorCode?: string;
	    isRetrying: boolean;
	    maxIntervalReached: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RetryState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.attemptNumber = source["attemptNumber"];
	        this.nextRetryIn = source["nextRetryIn"];
	        this.nextRetryAt = this.convertValues(source["nextRetryAt"], null);
	        this.lastError = source["lastError"];
	        this.lastErrorCode = source["lastErrorCode"];
	        this.isRetrying = source["isRetrying"];
	        this.maxIntervalReached = source["maxIntervalReached"];
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

}

export namespace version {
	
	export class Info {
	    version: string;
	    commit: string;
	    buildDate: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.commit = source["commit"];
	        this.buildDate = source["buildDate"];
	    }
	}
	export class UpdateInfo {
	    available: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    releaseUrl: string;
	    releaseNotes: string;
	    publishedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.releaseUrl = source["releaseUrl"];
	        this.releaseNotes = source["releaseNotes"];
	        this.publishedAt = source["publishedAt"];
	    }
	}

}

