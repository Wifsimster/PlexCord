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
	    serverUrl: string;
	    userId: string;
	    userName: string;
	    connected: boolean;
	    polling: boolean;
	    inErrorState: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PlexConnectionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serverUrl = source["serverUrl"];
	        this.userId = source["userId"];
	        this.userName = source["userName"];
	        this.connected = source["connected"];
	        this.polling = source["polling"];
	        this.inErrorState = source["inErrorState"];
	    }
	}
	export class ResourceStats {
	    timestamp: string;
	    memoryAllocMB: number;
	    memoryTotalMB: number;
	    goroutineCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ResourceStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.memoryAllocMB = source["memoryAllocMB"];
	        this.memoryTotalMB = source["memoryTotalMB"];
	        this.goroutineCount = source["goroutineCount"];
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
	    version: string;
	    isLocal: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Server(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.address = source["address"];
	        this.port = source["port"];
	        this.version = source["version"];
	        this.isLocal = source["isLocal"];
	    }
	}
	export class ValidationResult {
	    serverName: string;
	    serverVersion: string;
	    machineIdentifier: string;
	    libraryCount: number;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serverName = source["serverName"];
	        this.serverVersion = source["serverVersion"];
	        this.machineIdentifier = source["machineIdentifier"];
	        this.libraryCount = source["libraryCount"];
	        this.success = source["success"];
	    }
	}

}

export namespace retry {
	
	export class RetryState {
	    // Go type: time
	    nextRetryAt: any;
	    lastError?: string;
	    lastErrorCode?: string;
	    nextRetryIn: number;
	    attemptNumber: number;
	    isRetrying: boolean;
	    maxIntervalReached: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RetryState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nextRetryAt = this.convertValues(source["nextRetryAt"], null);
	        this.lastError = source["lastError"];
	        this.lastErrorCode = source["lastErrorCode"];
	        this.nextRetryIn = source["nextRetryIn"];
	        this.attemptNumber = source["attemptNumber"];
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
	    currentVersion: string;
	    latestVersion: string;
	    releaseUrl: string;
	    releaseNotes: string;
	    publishedAt: string;
	    available: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.releaseUrl = source["releaseUrl"];
	        this.releaseNotes = source["releaseNotes"];
	        this.publishedAt = source["publishedAt"];
	        this.available = source["available"];
	    }
	}

}

