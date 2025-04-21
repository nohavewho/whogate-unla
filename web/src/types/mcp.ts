import { z } from 'zod';

export const ToolSchema = z
  .object({
    name: z.string(),
    description: z.optional(z.string()),
    inputSchema: z
      .object({
        type: z.literal("object"),
        properties: z.optional(z.object({}).passthrough()),
      })
      .passthrough(),
  })
  .passthrough();

export type Tool = z.infer<typeof ToolSchema>;

export const ListToolsResultSchema = z.object({
  tools: z.array(ToolSchema),
});

export type ListToolsResult = z.infer<typeof ListToolsResultSchema>; 