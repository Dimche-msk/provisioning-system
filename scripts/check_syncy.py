import socket
import uuid
import random

# Настройки
TARGET_IP = "192.168.88.72"
TARGET_PORT = 5060
MY_IP = "192.168.88.111"  # Замените на IP вашего компьютера

def send_check_sync():
    call_id = f"{uuid.uuid4()}@{MY_IP}"
    from_tag = "".join(random.choices("0123456789abcdef", k=8))
    
    # Формируем SIP NOTIFY пакет
    # Event: check-sync — стандарт для большинства телефонов
    # Некоторые модели Mitel также понимают Event: resync
    sip_packet = (
        f"NOTIFY sip:{TARGET_IP} SIP/2.0\r\n"
        f"Via: SIP/2.0/UDP {MY_IP}:5060;branch=z9hG4bK-{uuid.uuid4()}\r\n"
        f"From: <sip:provisioning@{MY_IP}>;tag={from_tag}\r\n"
        f"To: <sip:{TARGET_IP}>\r\n"
        f"Call-ID: {call_id}\r\n"
        f"CSeq: 1 NOTIFY\r\n"
        f"Contact: <sip:provisioning@{MY_IP}>\r\n"
        f"Event: check-sync;reboot=true\r\n"
        f"Max-Forwards: 70\r\n"
        f"Subscription-State: terminated\r\n"
        f"Content-Length: 0\r\n"
        f"\r\n"
    )

    print(f"Sending NOTIFY to {TARGET_IP}:{TARGET_PORT}...")
    print(sip_packet)

    # Отправляем через UDP
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    try:
        sock.sendto(sip_packet.encode(), (TARGET_IP, TARGET_PORT))
        print("Packet sent successfully.")
    except Exception as e:
        print(f"Error: {e}")
    finally:
        sock.close()

if __name__ == "__main__":
    send_check_sync()

