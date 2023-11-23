// services/hooksService.ts
import { redisClient } from "../utils/redis";
import { HookModel } from "../models/hookModel";
import { strVal, strValOrUndef } from "@paroi/data-formatters-lib";
import { appLog } from "../share/app-log";

const streamKey = "hookStream";

export const startHooksListener = async () => {
  await redisClient.xgroup(
    "CREATECONSUMER",
    streamKey,
    "sendhooks-group",
    "$",
    function (err) {
      if (err) appLog.error("Error creating consumer group:", err);
    }
  );

  await redisClient.xreadgroup(
    "GROUP",
    "sendhooks-group",
    "consumer-1",
    "COUNT",
    0,
    "BLOCK",
    1,
    "STREAMS",
    streamKey,
    ">",
    (err, streams) => {
      if (err) {
        appLog.error("Error reading stream:", err);
      } else if (streams) {
        appLog.debug("streams", streams);
        const [stream] = streams;

        appLog.debug("streamData", JSON.stringify(stream, null, 2));
        const [streamName, messages] = stream as any;
        const [message] = messages;
        const [id, data] = message;
        const hookData = JSON.parse(data[1]);
        appLog.debug("hookData", hookData);
        handleHookCreation(id, hookData);
      }
      startHooksListener(); // Continue listening for more messages
    }
  );
};

const handleHookCreation = async (id: string, hookData: any): Promise<void> => {
  try {
    const hook = await HookModel.create({
      id,
      status: strVal(hookData.status),
      created: new Date(strVal(hookData.created)).getTime(),
      error: strValOrUndef(hookData.error),
    });

    appLog.debug("MONGO HOOK", hook);
  } catch (error) {
    appLog.error("Error creating hook:", error);
  }
};
