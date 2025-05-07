import {
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Button,
} from "@heroui/react";

interface WechatQRCodeProps {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
}

export function WechatQRCode({ isOpen, onOpenChange }: WechatQRCodeProps) {
  return (
    <Modal isOpen={isOpen} onOpenChange={onOpenChange} size="sm">
      <ModalContent>
        <ModalHeader>加入微信社区群</ModalHeader>
        <ModalBody>
          <div className="flex flex-col items-center justify-center">
            <img
              src="/wechat-qrcode.png"
              alt="WeChat QR Code"
              className="w-64 h-64 object-contain"
            />
            <p className="mt-4 text-center text-muted-foreground">
              扫描二维码添加微信
            </p>
            <p className="mt-4 text-center text-muted-foreground">
              备注mcp-gateway或mcpgw
            </p>
          </div>
        </ModalBody>
        <ModalFooter>
          <Button color="primary" onPress={() => onOpenChange(false)}>
            关闭
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
}
