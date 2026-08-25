import SignupForm from "../components/auth/SignupForm";
import AuthCard from "../components/layout/AuthCard";

export default function SignupPage() {
  return (
    <AuthCard title="Create an account" subtitle="Start in under a minute.">
      <SignupForm />
    </AuthCard>
  );
}
