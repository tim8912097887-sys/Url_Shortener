import { Link, useNavigate } from "react-router";
import { useForm } from "../../hooks/useForm";
import { signupSchema, type SignupSchemaType } from "../../schema/signup";
import { useAuthStore } from "../../store/useAuthStore";
import { ValidateInput } from "../../utils/validation";
import Form from "../ui/form/Form";
import Button from "../ui/button/Button";
import Alert from "../ui/alert/Alert";
import Divider from "../ui/divider";
import OAuthButtons from "./OAuthButtons";

export default function SignupForm() {
  const navigate = useNavigate();
  const signup = useAuthStore((s) => s.signup);
  const oauthLogin = useAuthStore((s) => s.oauthLogin);
  const submittingError = useAuthStore((s) => s.submittingError);
  const redirectTo = "/";
  const handleOAuthLogin = async () => {
    const data = await oauthLogin();
    console.log("Oauth data", data);
    if (data) {
      navigate(redirectTo, { replace: true });
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
  } = useForm<SignupSchemaType>({
    initialValues: { username: "", email: "", password: "" },
    onSubmit: async (formValues) => {
      await signup(formValues);
      // Set state for login page to show "just signed up" message
      navigate("/login", { replace: true, state: { justSignedUp: true } });
    },
    validate: (values: SignupSchemaType) => {
      const result = ValidateInput(values, signupSchema);
      return result;
    },
  });

  return (
    <div className="space-y-6">
      <Form onSubmit={handleSubmit}>
        {submitError && <Alert variant="error">{submitError}</Alert>}
        {submittingError && <Alert variant="error">{submittingError}</Alert>}
        <Form.Group
          controlId="signup-username"
          isInvalid={touched.username && Boolean(errors.username)}
        >
          <Form.Label>Username</Form.Label>
          <Form.Control
            type="text"
            name="username"
            autoComplete="username"
            placeholder="jane_doe"
            value={values.username}
            onChange={handleChange}
            onBlur={handleBlur}
          />
          <Form.Feedback>{touched.username && errors.username}</Form.Feedback>
        </Form.Group>

        <Form.Group
          controlId="signup-email"
          isInvalid={touched.email && Boolean(errors.email)}
        >
          <Form.Label>Email</Form.Label>
          <Form.Control
            type="email"
            name="email"
            autoComplete="email"
            placeholder="you@example.com"
            value={values.email}
            onChange={handleChange}
            onBlur={handleBlur}
          />
          <Form.Feedback>{touched.email && errors.email}</Form.Feedback>
        </Form.Group>

        <Form.Group
          controlId="signup-password"
          isInvalid={touched.password && Boolean(errors.password)}
        >
          <Form.Label>Password</Form.Label>
          <Form.Control
            type="password"
            name="password"
            autoComplete="new-password"
            placeholder="At least 8 characters"
            value={values.password}
            onChange={handleChange}
            onBlur={handleBlur}
          />
          <Form.Feedback>{touched.password && errors.password}</Form.Feedback>
        </Form.Group>

        <Form.Group
          controlId="signup-confirm-password"
          isInvalid={touched.confirmPassword && Boolean(errors.confirmPassword)}
        >
          <Form.Label>Confirm password</Form.Label>
          <Form.Control
            type="password"
            name="confirmPassword"
            autoComplete="new-password"
            placeholder="Re-enter your password"
            value={values.confirmPassword}
            onChange={handleChange}
            onBlur={handleBlur}
          />
          <Form.Feedback>
            {touched.confirmPassword && errors.confirmPassword}
          </Form.Feedback>
        </Form.Group>

        <Button type="submit" fullWidth isLoading={isSubmitting}>
          Create account
        </Button>
      </Form>

      <Divider label="or" />
      <OAuthButtons onOAuth={handleOAuthLogin} />

      <p className="text-center text-sm text-slate-500">
        Already have an account?{" "}
        <Link
          to="/login"
          className="font-medium text-teal-600 hover:text-teal-700"
        >
          Log in
        </Link>
      </p>
    </div>
  );
}
