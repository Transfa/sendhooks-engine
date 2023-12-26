import Redis from "ioredis";
import { appConfig } from "../configuration";

export const redisClient = new Redis(appConfig.redis);
