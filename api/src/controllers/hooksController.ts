import { Request, Response } from "express";
import { HookModel } from "../models/hookModel";
import { appLog } from "../share/app-log";

export class HookController {
  static async findAll(_: Request, res: Response) {
    try {
      const hooks = await HookModel.find();
      return res.json(hooks);
    } catch (error) {
      appLog.error("Error fetching hooks:", error);
      return res
        .status(500)
        .json({ code: "SERVER_ERROR", description: "Internal Server Error" });
    }
  }

  static async findOne(req: Request, res: Response) {
    const { hookId
    } = req.params;
    try {
      const hook = await HookModel.findById(hookId);
      if (!hook) {
        return res
          .status(404)
          .json({ code: "HOOK_NOT_FOUND", description: "Hook not found" });
      }

      return res.json(hook);
    } catch (error) {
      appLog.error("Error fetching hook:", error);

      return res
        .status(500)
        .json({ code: "SERVER_ERROR", description: "Internal Server Error" });
    }
  }
}
