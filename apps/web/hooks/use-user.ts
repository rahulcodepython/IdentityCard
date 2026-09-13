import { authClient } from "@/lib/auth-client";
import { ROUTE_LOGIN } from "@/lib/constants";
import { useSessionStore } from "@/store/session.store";
import { useRouter } from "next/navigation";

const useUser = () => {
    const user = useSessionStore((state) => state.user);
    const reset = useSessionStore((state) => state.setUnauthenticated);
    const activeOrgId = useSessionStore((state) => state.activeOrgId);

    const router = useRouter();

    const handleLogout = async () => {
        await authClient
            .signOut()
            .then(() => {
                reset();
            })
            .then(() => {
                router.push(ROUTE_LOGIN);
            })
            .catch((err) => {
                console.error("Logout failed:", err);
            });
    };

    if (!user) {
        handleLogout();
    }

    return { handleLogout, user, activeOrgId };
};

export default useUser;