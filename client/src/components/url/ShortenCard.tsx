import type { ApiError } from "../../api/errors/api-error";
import { urlService } from "../../api/services/url.service";
import { env } from "../../config/env";
import { useForm } from "../../hooks/useForm";
import { urlSchema, type UrlSchemaType } from "../../schema/url";
import { ValidateInput } from "../../utils/validation";
import Alert from "../ui/alert";
import Button from "../ui/button";
import Form from "../ui/form/Form";

type ShortenCardProps = {
  setShortUrl: (shortUrl: string) => void;
};

const ShortenCard = ({ setShortUrl }: ShortenCardProps) => {
  const shortenUrl = async (payload: UrlSchemaType) => {
    try {
      const response = await urlService.shortenUrl(payload);
      return response;
    } catch (error: ApiError | any) {
      throw new Error(error.message || "An unexpected error occurred");
    }
  };

  const {
    values,
    errors,
    touched,
    isSubmitting,
    submitError,
    handleChange,
    handleBlur,
    handleSubmit,
  } = useForm({
    initialValues: { url: "" },
    validate: (values: UrlSchemaType) => {
      const result = ValidateInput(values, urlSchema);
      return result;
    },
    onSubmit: async (formValues) => {
      const response = await shortenUrl(formValues);
      setShortUrl(env.apiBaseUrl + "/urls/" + response.data.shortUrl);
    },
  });

  return (
    <div className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
      <div className="space-y-6">
        <Form onSubmit={handleSubmit}>
          {submitError && <Alert variant="error">{submitError}</Alert>}
          <Form.Group controlId="url" isInvalid={!!errors.url && touched.url}>
            <Form.Label>URL</Form.Label>
            <Form.Control
              type="url"
              placeholder="https://example.com"
              name="url"
              value={values.url}
              onChange={handleChange}
              onBlur={handleBlur}
            />
            <Form.Feedback>{touched.url && errors.url}</Form.Feedback>
          </Form.Group>
          <Button type="submit" fullWidth isLoading={isSubmitting}>
            Shorten
          </Button>
        </Form>
      </div>
    </div>
  );
};

export default ShortenCard;
