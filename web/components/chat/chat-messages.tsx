type ChatMessagesProps = React.ComponentPropsWithoutRef<"div">;

export default function ChatMessages({
  className,
  ...props
}: ChatMessagesProps) {
  return <div {...props} className={className} />;
}
