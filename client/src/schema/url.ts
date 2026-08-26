import z from "zod";

export const urlSchema = z.object({
  url: z
    .string({ error: "URL must be a string" })
    // Regex enforces protocol + domain + TLD
    .regex(
      /^(https?:\/\/)([\w.-]+)\.([a-z]{2,})([/\w%?&=.#-]*)?$/,
      "URL must start with http(s) and have a valid domain",
    )
    // Refine with URL parser for full validation
    .refine(
      (val) => {
        try {
          new URL(val);
          return true;
        } catch {
          return false;
        }
      },
      { message: "Invalid URL structure" },
    ),
});

export type UrlSchemaType = z.infer<typeof urlSchema>;
