import z from "zod";

export const urlSchema = z.object({
  url: z
    .string("URL must be a string")
    .regex(
      /^(https?:\/\/)([\w.-]+)\.([a-z]{2,})([/\w .-]*\/?)?$/,
      "URL must be a valid URL",
    ),
});

export type UrlSchemaType = z.infer<typeof urlSchema>;
