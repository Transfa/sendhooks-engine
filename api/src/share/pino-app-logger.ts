import { EventEmitter } from "events";
import pino from "pino";
import { promiseToHandle } from '@paroi/async-lib'

const levels: AppLogLevel[] = ["error", "warn", "info", "debug", "trace"];
const levelIndexes = new Map(levels.map((l, index) => [l, index]));

export interface AppLog {
	error(...messages: any[]): void;
	warn(...messages: any[]): void;
	info(...messages: any[]): void;
	debug(...messages: any[]): void;
	trace(...messages: any[]): void;
	on(level: AppLogLevel, listener: (ev: AppLogEvent) => void): void;
	untilReady: Promise<void>;
	flushSync(): void;
}

export interface AppLogEvent {
	level: AppLogLevel;
	textMessage: string;
	originalMessages: any[];
}

export type AppLogLevel = Exclude<keyof AppLog, "flushSync" | "untilReady" | "on">;

export interface AppLogOptions {
	level: "silent" | AppLogLevel;
	/**
	 * Omit for stdout.
	 */
	file?: string;
	prettyPrint?: boolean;
}

/**
 * Warning: install pino-pretty in development mode only.
 */
export function createAppLog({ file, level, prettyPrint }: AppLogOptions): AppLog {
	const destination = pino.destination({
		dest: file,
		// minLength: 4096,
		sync: false,
	});

	const emitter = new EventEmitter();
	let ready = false;
	const { promise: untilReady, resolve, reject } = promiseToHandle<void>();
	destination.on("error", reject);
	destination.on("ready", () => {
		ready = true;
		destination.off("error", reject);
		destination.on("error", (error: unknown) => console.error("[Error in Pino]", error));
		resolve();
	});

	let waitingMessages: any[][] | undefined;

	function makeLogFn(level: AppLogLevel) {
		return (...messages: any[]) => {
			const textMessage = messagesToString(messages);
			if (ready) logger[level](textMessage);
			else {
				if (!waitingMessages) {
					waitingMessages = [];
					console.warn("There is something to log before the logger is ready");
					void untilReady.then(() => {
						if (waitingMessages) {
							if (file)
								waitingMessages.forEach((wMessages) => logger[level](messagesToString(wMessages)));
							waitingMessages = undefined;
						}
					});
				}
				// eslint-disable-next-line @typescript-eslint/no-unsafe-call
				if (level in console) (console as any)[level](...messages);
				else console.log(...messages);
				messages.unshift("[DELAYED]");
				waitingMessages.push(messages);
			}
			try {
				emitAppLogEvent(
					{
						level,
						textMessage,
						originalMessages: messages,
					},
					emitter,
				);
			} catch (error) {
				if (ready) logger.error(messagesToString(["Error in app log listener:", error]));
				else console.error("Error in app log listener:", error);
			}
		};
	}

	const logger = pino(
		{
			level,
			customLevels: {
				special: 55,
				alive: 25,
			},
			prettyPrint: prettyPrint
				? {
						translateTime: "yyyy-mm-dd HH:MM:ss.l",
						ignore: "hostname,pid",
				  }
				: undefined,
		},
		destination,
	);
	console.log(`Application log with level '${level}' is in: ${file ?? "stdout"}.`);

	return {
		error: makeLogFn("error"),
		warn: makeLogFn("warn"),
		info: makeLogFn("info"),
		debug: makeLogFn("debug"),
		trace: makeLogFn("trace"),
		on: (level, listener) => emitter.on(`on-${level}`, listener),
		untilReady,
		flushSync() {
			if (ready) destination.flushSync();
			else console.warn("Flush is called before the logger is ready.");
		},
	};
}

function messagesToString(messages: unknown[]): string {
	return messages.map((msg) => messageToString(msg)).join(" ");
}

function messageToString(msg: unknown, parents: unknown[] = []): string {
	if (parents.includes(msg)) return "[recursive-ref]";
	if (parents.length > 5) return "[too-deep]";
	switch (typeof msg) {
		case "string":
			return msg;
		case "number":
		case "bigint":
		case "boolean":
		case "undefined":
		case "symbol":
			return String(msg);
		case "function":
			return `[function ${msg.name}]`;
		case "object":
			if (msg === null) return "null";
			if (Array.isArray(msg))
				return `[${msg.map((child) => messageToString(child, [...parents, msg])).join(",")}]`;
			if (msg instanceof Error) return msg.stack ?? `Error: ${msg.message ?? "-no-message-"}`;
			return `{${Object.entries(msg)
				.map(([key, child]) => `${key}: ${messageToString(child, [...parents, msg])}`)
				.join(",")}}`;
		default:
			return `Unexpected message type: ${typeof msg}`;
	}
}

function emitAppLogEvent(ev: AppLogEvent, emitter: EventEmitter) {
	let index = levelIndexes.get(ev.level);
	if (index === undefined) throw new Error(`Invalid level '${ev.level}'`);
	do {
		emitter.emit(`on-${levels[index]}`, ev);
	} while (++index < levels.length);
}
